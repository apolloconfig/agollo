package dto

type ApolloConfig struct {
	appId string
	cluster string
	namespaceName string
	configurations map[string]string
	releaseKey string
}

func CreateApolloConfig(appId string,
	cluster string,
	namespaceName string,
	releaseKey string) (apolloConfig *ApolloConfig) {
	apolloConfig=&ApolloConfig{
		appId:appId,
		cluster:cluster,
		namespaceName:namespaceName,
		releaseKey:releaseKey,
	}
	return
}