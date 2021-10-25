package config

var base = mergeConfig(
	fileLocationConfig,
	logLevelConfig,
	notionDatabaseConfig,
	notionTokenConfig,
	repositoryNameConfig,
	repositoryPullRequestConfig,
	scanTypeConfig,
	slackConfig,
)
