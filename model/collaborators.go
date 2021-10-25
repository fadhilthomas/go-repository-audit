package model

type Collaborators struct {
	OrganizationName string
	RepositoryName   string
	RepositoryOwner  string
	UserName         string
	UserLogin        string
	Permission       map[string]bool
}
