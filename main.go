package main

import (
	"context"
	"github.com/fadhilthomas/go-repository-audit/config"
	"github.com/fadhilthomas/go-repository-audit/model"
	"github.com/google/go-github/v39/github"
	"github.com/jomei/notionapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/oauth2"
)

var (
	notionPageList []notionapi.Page
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

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	repositoryListByOrgOptions := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	listCollaboratorsOptions := &github.ListCollaboratorsOptions{}

	var githubRepositoryList []model.GitHubRepository

	// get all pages of results
	for {
		repositoryList, resp, err := client.Repositories.ListByOrg(ctx, organizationName, repositoryListByOrgOptions)
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
			continue
		}
		if resp.StatusCode != 200 {
			continue
		}

		for _, repository := range repositoryList {
			repositoryName := *repository.Name
			repositoryOwner := *repository.Owner.Login
			userList, resp, err := client.Repositories.ListCollaborators(ctx, repositoryOwner, repositoryName, listCollaboratorsOptions)

			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				continue
			}
			if resp.StatusCode != 200 {
				continue
			}

			for _, user := range userList {
				githubRepository := model.GitHubRepository{}
				githubRepository.OrganizationName = organizationName
				githubRepository.RepositoryName = repositoryName
				githubRepository.RepositoryOwner = repositoryOwner
				githubRepository.UserLogin = *user.Login
				githubRepository.Permission = user.Permissions
				githubRepositoryList = append(githubRepositoryList, githubRepository)
				_, err = model.InsertNotionRepository(notionDatabase, githubRepository)
				if err != nil {
					log.Error().Stack().Err(err).Msg("")
					return
				}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		repositoryListByOrgOptions.Page = resp.NextPage
	}
}
