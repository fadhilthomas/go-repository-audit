package model

import (
	"context"
	"github.com/fadhilthomas/go-repository-audit/config"
	"github.com/jomei/notionapi"
	"github.com/pkg/errors"
)

func OpenNotionDB() (client *notionapi.Client) {
	notionToken := config.GetStr(config.NOTION_TOKEN)
	client = notionapi.NewClient(notionapi.Token(notionToken))
	return client
}

func QueryNotionRepositoryName(client *notionapi.Client, repositoryName string) (output []notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_DATABASE)

	databaseQueryRequest := &notionapi.DatabaseQueryRequest{
		PropertyFilter: &notionapi.PropertyFilter{
			Property: "Repository Name",
			Select: &notionapi.SelectFilterCondition{
				Equals: repositoryName,
			},
		},
	}

	res, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseId), databaseQueryRequest)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res.Results, nil
}

func InsertNotionRepository(client *notionapi.Client, repository GitHubRepository) (output *notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_DATABASE)

	pageInsertQuery := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseId),
		},
		Properties: notionapi.Properties{
			"Organization Name": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: repository.OrganizationName,
						},
					},
				},
			},
			"Repository Name": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.RepositoryName,
				},
			},
			"Owner": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.RepositoryOwner,
				},
			},
			"User Login": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.UserLogin,
				},
			},
			"Admin Permission": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["admin"],
			},
			"Maintain Permission": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["maintain"],
			},
			"Pull Permission": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["pull"],
			},
			"Push Permission": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["push"],
			},
		},
	}

	res, err := client.Page.Create(context.Background(), pageInsertQuery)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res, nil
}
