package config

var base = mergeConfig(
	fileLocationConfig,
	logLevelConfig,
	githubTokenConfig,
	organizationNameConfig,
)
