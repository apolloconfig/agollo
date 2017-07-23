package agollo

import (
	"testing"
	"os"
	"strconv"
	"time"
)

func TestInitRefreshInterval_1(t *testing.T) {
	os.Setenv(REFRESH_INTERVAL_KEY,"joe")

	err:=initRefreshInterval()
	NotNil(t,err)

	interval:="3"
	os.Setenv(REFRESH_INTERVAL_KEY,interval)
	err=initRefreshInterval()
	Nil(t,err)
	i,_:=strconv.Atoi(interval)
	Equal(t,time.Duration(i),REFRESH_INTERVAL)

}