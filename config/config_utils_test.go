package config

import (
	"testing"
	"os"
	"github.com/zouyx/agollo/test"
	"strconv"
)

func TestInitRefreshInterval(t *testing.T) {
	os.Setenv(REFRESH_INTERVAL_KEY,"joe")

	err:=initRefreshInterval()
	test.NotNil(t,err)

	interval:="3"
	os.Setenv(REFRESH_INTERVAL_KEY,interval)
	err=initRefreshInterval()
	test.Nil(t,err)
	i,_:=strconv.Atoi(interval)
	test.Equal(t,i,REFRESH_INTERVAL)

}