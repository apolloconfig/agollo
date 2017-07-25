package agollo

import (
	"testing"
	"os"
	"strconv"
	"time"
)

func TestInitRefreshInterval_1(t *testing.T) {
	os.Setenv(refresh_interval_key,"joe")

	err:=initRefreshInterval()
	NotNil(t,err)

	interval:="3"
	os.Setenv(refresh_interval_key,interval)
	err=initRefreshInterval()
	Nil(t,err)
	i,_:=strconv.Atoi(interval)
	Equal(t,time.Duration(i),refresh_interval)

}