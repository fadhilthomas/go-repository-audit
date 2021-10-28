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

	repositoryListByOrgOptions := &github.RepositoryListByOrgOptions{}

	collaboratorsListOptions := &github.ListCollaboratorsOptions{}

	// get all pages of results
	for {
		repositoryList, repositoryResp, err := client.Repositories.ListByOrg(ctx, organizationName, repositoryListByOrgOptions)
		if err != nil {
			log.Error().Stack().Err(errors.New(err.Error())).Msg("")
			continue
		}
		if repositoryResp.StatusCode != 200 {
			continue
		}

		for _, repository := range repositoryList {
			repositoryName := *repository.Name
			repositoryOwner := *repository.Owner.Login

			// get all collaborator
			for {
				collaboratorsList, collaboratorsResp, err := client.Repositories.ListCollaborators(ctx, repositoryOwner, repositoryName, collaboratorsListOptions)

				if err != nil {
					log.Error().Stack().Err(errors.New(err.Error())).Msg("")
					continue
				}

				if collaboratorsResp.StatusCode != 200 {
					continue
				}

				// get all page with repository name
				rl.Take()
				githubRepositoryNotionList, err := model.QueryNotionRepository(notionDatabase, repositoryName)
				if err != nil {
					log.Error().Stack().Err(err).Msg("")
					return
				}

				// if list of repository name page not empty
				// update all status to close
				if len(githubRepositoryNotionList) > 0 {
					for _, githubRepositoryNotionPage := range githubRepositoryNotionList {
						rl.Take()
						_, err = model.UpdateNotionRepositoryStatus(notionDatabase, githubRepositoryNotionPage.ID.String(), "close")
						if err != nil {
							log.Error().Stack().Err(err).Msg("")
							return
						}
					}
				}

				for _, user := range collaboratorsList {
					githubRepository := model.GitHubRepository{}
					githubRepository.OrganizationName = organizationName
					githubRepository.RepositoryName = repositoryName
					githubRepository.RepositoryOwner = repositoryOwner
					githubRepository.UserLogin = *user.Login
					githubRepository.Permission = user.Permissions

					rl.Take()
					// get page with repository name and user
					githubRepositoryUserNotion, err := model.QueryNotionRepositoryUser(notionDatabase, repositoryName, *user.Login)
					if err != nil {
						log.Error().Stack().Err(err).Msg("")
						return
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
							return
						}
					}
				}
				if collaboratorsResp.NextPage == 0 {
					break
				}
				collaboratorsListOptions.Page = collaboratorsResp.NextPage
			}
		}
		if repositoryResp.NextPage == 0 {
			break
		}
		repositoryListByOrgOptions.Page = repositoryResp.NextPage
	}
}
