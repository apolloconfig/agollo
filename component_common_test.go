package agollo

import (
	"testing"
	"github.com/zouyx/agollo/test"
)

func TestCreateApolloConfigWithJson(t *testing.T) {
	jsonStr := `{
  "appId": "100004458",
  "cluster": "default",
  "namespaceName": "application",
  "configurations": {
    "key1":"value1",
    "key2":"value2"
  },
  "releaseKey": "20170430092936-dee2d58e74515ff3"
}`

	config,err:=createApolloConfigWithJson([]byte(jsonStr))

	test.Nil(t,err)
	test.NotNil(t,config)

	test.Equal(t,"100004458",config.AppId)
	test.Equal(t,"default",config.Cluster)
	test.Equal(t,"application",config.NamespaceName)
	test.Equal(t,"20170430092936-dee2d58e74515ff3",config.ReleaseKey)
	test.Equal(t,"value1",config.Configurations["key1"])
	test.Equal(t,"value2",config.Configurations["key2"])

}

func TestCreateApolloConfigWithJsonError(t *testing.T) {
	jsonStr := `jklasdjflasjdfa`

	config,err:=createApolloConfigWithJson([]byte(jsonStr))

	test.NotNil(t,err)
	test.Nil(t,config)
}
