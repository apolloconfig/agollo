package namespaces

import (
	"github.com/qshuai/agollo/v4/env/config"
)

type Fetcher struct{}

func (f Fetcher) Fetch(c *config.AppConfig) ([]string, error) {
	var namespace []string
	err := SyncNamespaceList(func() config.AppConfig {
		return *c
	}, func(ns string) error {
		namespace = append(namespace, ns)
		return nil
	})

	return namespace, err
}
