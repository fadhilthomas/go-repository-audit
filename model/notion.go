package model

import (
	"context"
	"github.com/fadhilthomas/go-repository-audit/config"
	"github.com/fadhilthomas/go-repository-audit/model"
	"github.com/jomei/notionapi"
	"github.com/pkg/errors"
	"strings"
)

func OpenNotionDB() (client *notionapi.Client) {
	notionToken := config.GetStr(config.NOTION_TOKEN)
	client = notionapi.NewClient(notionapi.Token(notionToken))
	return client
}

func QueryNotionRepositoryName(client *notionapi.Client, repository model.GitHubRepository) (output []notionapi.Page, err error) {
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

func InsertNotionRepository(client *notionapi.Client, ) (output *notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_DATABASE)

	var path string
	var detail string
	if strings.Contains(scanType, "dependency") {
		path = "Package"
		detail = "CVSS Score"
	} else {
		path = "File Location"
		detail = "Line Number"
	}

	pageInsertQuery := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseId),
		},
		Properties: notionapi.Properties{
			"Name": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: RepositoryName,
						},
					},
				},
			},
			path: notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: RepositoryPath,
						},
					},
				},
			},
			detail: notionapi.NumberProperty{
				Number: RepositoryDetail,
			},
			"Repository": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repositoryName,
				},
			},
			"Repository Pull Request": notionapi.RichTextProperty{
				RichText: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: repositoryPullRequest,
						},
					},
				},
			},
			"Status": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: "open",
				},
			},
		},
	}

	res, err := client.Page.Create(context.Background(), pageInsertQuery)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res, nil
}

func UpdateNotionRepositoryStatus(client *notionapi.Client, pageId string, status string) (output *notionapi.Page, err error) {
	pageUpdateQuery := &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Status": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: status,
				},
			},
		},
	}

	res, err := client.Page.Update(context.Background(), notionapi.PageID(pageId), pageUpdateQuery)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	return res, nil
}