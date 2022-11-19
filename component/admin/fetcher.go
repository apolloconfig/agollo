package admin

import (
	"github.com/qshuai/agollo/v4/env/config"
)

type Fetcher struct{}

func (f Fetcher) Fetch(c *config.AppConfig) ([]string, error) {
	srvs := services.Load()
	if srvs != nil {
		return srvs.([]string), nil
	}

	return SyncAdminService(func() config.AppConfig {
		return *c
	})
}
