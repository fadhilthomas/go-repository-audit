package main

import (
	"context"
	"github.com/fadhilthomas/go-repository-audit/config"
	"github.com/fadhilthomas/go-repository-audit/model"
	"github.com/google/go-github/v39/github"
	"github.com/jomei/notionapi"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"go.uber.org/ratelimit"
	"golang.org/x/oauth2"
)

var (
	notionDatabase *notionapi.Client
)

func main() {
	config.Set(config.LOG_LEVEL, "info")
	if config.GetStr(config.LOG_LEVEL) == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	organizationName := config.GetStr(config.ORGANIZATION_NAME)
	githubToken := config.GetStr(config.GITHUB_TOKEN)

	notionDatabase = model.OpenNotionDB()
	rl := ratelimit.New(1)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	teamListByOrgOptions := &github.ListOptions{}

	var teamList []*github.Team
	var repoList []*github.Repository

	// get all teams by org name
	for {
		rl.Take()
		teamListRes, teamResp, err := client.Teams.ListTeams(ctx, organizationName, teamListByOrgOptions)
		if err != nil {
			log.Error().Stack().Err(errors.New(err.Error())).Msg("")
			continue
		}
		teamList = append(teamList, teamListRes...)

		if teamResp.NextPage == 0 {
			break
		}
		teamListByOrgOptions.Page = teamResp.NextPage
	}

	for _, teamSlug := range teamList {
		repoListByTeamOptions := &github.ListOptions{}

		for {
			rl.Take()

			// get all repos by team slug
			repoListRes, repoResp, err := client.Teams.ListTeamReposBySlug(ctx, organizationName, *teamSlug.Slug, repoListByTeamOptions)
			if err != nil {
				log.Error().Stack().Err(errors.New(err.Error())).Msg("")
				continue
			}
			repoList = append(repoList, repoListRes...)

			if repoResp.NextPage == 0 {
				break
			}
			repoListByTeamOptions.Page = repoResp.NextPage
		}
	}

	for _, repo := range removeDuplicate(repoList) {

		collaboratorsListOptions := &github.ListCollaboratorsOptions{}

		var userList []*github.User
		repositoryName := *repo.Name
		repositoryOwner := *repo.Owner.Login

		for {
			userListRes, collaboratorsResp, err := client.Repositories.ListCollaborators(ctx, repositoryOwner, repositoryName, collaboratorsListOptions)

			if err != nil {
				log.Error().Stack().Err(errors.New(err.Error())).Msg("")
				continue
			}
			userList = append(userList, userListRes...)

			if collaboratorsResp.NextPage == 0 {
				break
			}
			collaboratorsListOptions.Page = collaboratorsResp.NextPage
		}

		// get all page with repo name and open status
		rl.Take()
		githubRepositoryNotionList, err := model.QueryNotionRepositoryStatus(notionDatabase, repositoryName, "open")
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
			continue
		}

		// if list of pages not empty
		// update all status to close
		if len(githubRepositoryNotionList) > 0 {
			for _, githubRepositoryNotionPage := range githubRepositoryNotionList {
				rl.Take()
				_, err = model.UpdateNotionRepositoryStatus(notionDatabase, githubRepositoryNotionPage.ID.String(), "close")
				if err != nil {
					log.Error().Stack().Err(err).Msg("")
					continue
				}
			}
		}

		for _, user := range userList {
			githubRepository := model.GitHubRepository{}
			githubRepository.OrganizationName = organizationName
			githubRepository.RepositoryName = repositoryName
			githubRepository.RepositoryOwner = repositoryOwner
			githubRepository.UserLogin = *user.Login
			githubRepository.Permission = user.Permissions

			if githubRepository.Permission["push"] {
				rl.Take()
				// get page with repository name and user
				githubRepositoryUserNotion, err := model.QueryNotionRepositoryUser(notionDatabase, repositoryName, *user.Login)
				if err != nil {
					log.Error().Stack().Err(err).Msg("")
					continue
				}

				// if list of repository name and user page empty
				// insert to notion
				if len(githubRepositoryUserNotion) == 0 {
					rl.Take()
					_, err = model.InsertNotionRepository(notionDatabase, "report-log", githubRepository)
					if err != nil {
						log.Error().Stack().Err(err).Msg("")
						continue
					}

					rl.Take()
					_, err = model.InsertNotionRepository(notionDatabase, "change-log", githubRepository)
					if err != nil {
						log.Error().Stack().Err(err).Msg("")
						continue
					}
				} else {
					rl.Take()
					_, err = model.UpdateNotionRepository(notionDatabase, githubRepositoryUserNotion[0].ID.String(), githubRepository, "open")
					if err != nil {
						log.Error().Stack().Err(err).Msg("")
						continue
					}
				}
			}
		}

	}
}

func removeDuplicate(duplicate []*github.Repository) []github.Repository {
	var unique []github.Repository
	type key struct{ value1 int64 }
	m := make(map[key]int)
	for _, v := range duplicate {
		k := key{*v.ID}
		if i, ok := m[k]; ok {
			unique[i] = *v
		} else {
			m[k] = len(unique)
			unique = append(unique, *v)
		}
	}
	return unique
}
