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

func QueryNotionRepositoryUser(client *notionapi.Client, repositoryName string, userLogin string) (output []notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_REPORT_DATABASE)

	var pageList []notionapi.Page
	var cursor notionapi.Cursor
	for hasMore := true; hasMore; {
		databaseQueryRequest := &notionapi.DatabaseQueryRequest{
			CompoundFilter: &notionapi.CompoundFilter{
				notionapi.FilterOperatorAND: []notionapi.PropertyFilter{
					{
						Property: "Repository",
						Select: &notionapi.SelectFilterCondition{
							Equals: repositoryName,
						},
					},
					{
						Property: "User",
						Select: &notionapi.SelectFilterCondition{
							Equals: userLogin,
						},
					},
				},
			},
			StartCursor: cursor,
		}
		resp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseId), databaseQueryRequest)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		pageList = append(pageList, resp.Results...)
		hasMore = resp.HasMore
		cursor = resp.NextCursor
	}
	return pageList, nil
}

func QueryNotionRepositoryStatus(client *notionapi.Client, repositoryName string, status string) (output []notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_REPORT_DATABASE)

	var pageList []notionapi.Page
	var cursor notionapi.Cursor
	for hasMore := true; hasMore; {
		databaseQueryRequest := &notionapi.DatabaseQueryRequest{
			CompoundFilter: &notionapi.CompoundFilter{
				notionapi.FilterOperatorAND: []notionapi.PropertyFilter{
					{
						Property: "Repository",
						Select: &notionapi.SelectFilterCondition{
							Equals: repositoryName,
						},
					},
					{
						Property: "Status",
						Select: &notionapi.SelectFilterCondition{
							Equals: status,
						},
					},
				},
			},
			StartCursor: cursor,
		}
		resp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseId), databaseQueryRequest)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		pageList = append(pageList, resp.Results...)
		hasMore = resp.HasMore
		cursor = resp.NextCursor
	}
	return pageList, nil
}

func InsertNotionNewPermission(client *notionapi.Client, notionDatabaseType string, repository GitHubRepository) (output *notionapi.Page, err error) {
	var databaseId string
	if notionDatabaseType == "change-log" {
		databaseId = config.GetStr(config.NOTION_PERMISSION_DATABASE)
	} else if notionDatabaseType == "report-log" {
		databaseId = config.GetStr(config.NOTION_REPORT_DATABASE)
	}

	pageInsertQuery := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseId),
		},
		Properties: notionapi.Properties{
			"Organization": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: repository.OrganizationName,
						},
					},
				},
			},
			"Repository": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.RepositoryName,
				},
			},
			"User": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.UserLogin,
				},
			},
			"Admin": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["admin"],
			},
			"Push": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["push"],
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

func InsertNotionNewRepository(client *notionapi.Client, organizationName string, repositoryName string, repositoryOwner string) (output *notionapi.Page, err error) {
	databaseId := config.GetStr(config.NOTION_REPOSITORY_DATABASE)

	pageInsertQuery := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseId),
		},
		Properties: notionapi.Properties{
			"Organization": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: organizationName,
						},
					},
				},
			},
			"Repository": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repositoryName,
				},
			},
			"Owner": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repositoryOwner,
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

func UpdateNotionRepository(client *notionapi.Client, pageId string, repository GitHubRepository, status string) (output *notionapi.Page, err error) {
	pageUpdateQuery := &notionapi.PageUpdateRequest{
		Properties: notionapi.Properties{
			"Organization": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Text: notionapi.Text{
							Content: repository.OrganizationName,
						},
					},
				},
			},
			"Repository": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.RepositoryName,
				},
			},
			"User": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: repository.UserLogin,
				},
			},
			"Admin": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["admin"],
			},
			"Push": notionapi.CheckboxProperty{
				Checkbox: repository.Permission["push"],
			},
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
