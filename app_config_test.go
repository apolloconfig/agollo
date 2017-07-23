package agollo

import (
	"testing"
	"os"
	"strconv"
	"time"
	"github.com/zouyx/agollo/test"
)

func TestInitRefreshInterval_1(t *testing.T) {
	os.Setenv(refresh_interval_key,"joe")

	err:=initRefreshInterval()
	test.NotNil(t,err)

	interval:="3"
	os.Setenv(refresh_interval_key,interval)
	err=initRefreshInterval()
	test.Nil(t,err)
	i,_:=strconv.Atoi(interval)
	test.Equal(t,time.Duration(i),refresh_interval)

}