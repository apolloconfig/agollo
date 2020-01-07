package component

import (
	"fmt"
	"github.com/zouyx/agollo/v2/env/config"
	"net/url"

	"github.com/zouyx/agollo/v2/env"
	"github.com/zouyx/agollo/v2/utils"
)

type AbsComponent interface {
	Start()
}

func StartRefreshConfig(component AbsComponent) {
	component.Start()
}

func GetConfigURLSuffix(config *config.AppConfig, namespaceName string) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(namespaceName),
		url.QueryEscape(env.GetCurrentApolloConfigReleaseKey(namespaceName)),
		utils.GetInternal())
}
