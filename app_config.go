package agollo

import (
	"os"
	"strconv"
	"time"
	"fmt"
	"net/url"
	"github.com/cihub/seelog"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

const appConfigFileName  ="app.properties"

var (
	refresh_interval = 5 *time.Minute //5m
	refresh_interval_key = "apollo.refreshInterval"  //

	long_poll_interval = 5 *time.Second //5s
	long_poll_connect_timeout  = 1 * time.Minute //1m

	connect_timeout  = 1 * time.Second //1s
	//for on error retry
	on_error_retry_interval = 1 * time.Second //1s
	//for typed config cache of parser result, e.g. integer, double, long, etc.
	//max_config_cache_size    = 500             //500 cache key
	//config_cache_expire_time = 1 * time.Minute //1 minute

	//max retries connect apollo
	max_retries=5

	//refresh ip list
	refresh_ip_list_interval=20 *time.Minute //20m

	//appconfig
	appConfig *AppConfig

	//real servers ip
	servers map[string]*serverInfo=make(map[string]*serverInfo,0)

	//next try connect period - 60 second
	next_try_connect_period int64=60
)

type AppConfig struct {
	AppId string `json:"appId"`
	Cluster string `json:"cluster"`
	NamespaceName string `json:"namespaceName"`
	Ip string `json:"ip"`
	NextTryConnTime int64 `json:"-"`
}

func (this *AppConfig) getHost() string{
	return "http://"+this.Ip+"/"
}

//if this connect is fail will set this time
func (this *AppConfig) setNextTryConnTime(){
	this.NextTryConnTime=time.Now().Unix()+next_try_connect_period
}

//is connect by ip directly
//false : no
//true : yes
func (this *AppConfig) isConnectDirectly() bool{
	if this.NextTryConnTime==0||this.NextTryConnTime>time.Now().Unix(){
		return false
	}

	return true
}

func (this *AppConfig) selectHost() string{
	if this.isConnectDirectly(){
		return this.getHost()
	}




	return ""
}


type serverInfo struct {
	AppName string `json:"appName"`
	InstanceId string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`

}

func init() {
	//init common
	initCommon()

	//init config
	initConfig()
}

func initCommon()  {

	initRefreshInterval()
}

func initConfig() {
	var err error
	//init config file
	appConfig,err = loadJsonConfig(appConfigFileName)

	if err!=nil{
		panic(err)
	}

	go func(appConfig *AppConfig) {
		apolloConfig:=&ApolloConfig{}
		apolloConfig.AppId=appConfig.AppId
		apolloConfig.Cluster=appConfig.Cluster
		apolloConfig.NamespaceName=appConfig.NamespaceName

		updateApolloConfig(apolloConfig)
	}(appConfig)
}

//set timer for update ip list
//interval : 20m
func initServerIpList() {
	t2 := time.NewTimer(refresh_ip_list_interval)
	for {
		select {
		case <-t2.C:
			syncServerIpList()
			t2.Reset(refresh_ip_list_interval)
		}
	}
}

//sync ip list from server
//then
//1.update cache
//2.store in disk
func syncServerIpList() error{
	client := &http.Client{
		Timeout:connect_timeout,
	}

	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=getServicesConfigUrl(appConfig)
	seelog.Debug("url:",url)

	retry:=0
	var responseBody []byte
	var err error
	var res *http.Response
	for{
		retry++

		if retry>max_retries{
			break
		}

		res,err=client.Get(url)

		if res==nil||err!=nil{
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			continue
		}

		//not modified break
		switch res.StatusCode {
		case http.StatusOK:
			responseBody, err = ioutil.ReadAll(res.Body)
			if err!=nil{
				seelog.Error("Connect Apollo Server Fail,Error:",err)
				continue
			}

			tmpServerInfo:=make([]*serverInfo,0)

			err = json.Unmarshal(responseBody,&tmpServerInfo)

			if err!=nil{
				seelog.Error("Unmarshal json Fail,Error:",err)
				return err
			}

			if len(tmpServerInfo)==0 {
				seelog.Info("get no real server!")
				return nil
			}

			for _,server :=range tmpServerInfo {
				servers[server.InstanceId]=server
			}

			return nil
		default:
			seelog.Error("Connect Apollo Server Fail,Error:",err)
			if res!=nil{
				seelog.Error("Connect Apollo Server Fail,StatusCode:",res.StatusCode)
			}
			// if error then sleep
			time.Sleep(on_error_retry_interval)
			continue
		}
	}

	seelog.Debug(responseBody)

	seelog.Error("Over Max Retry Still Error,Error:",err)
	if err==nil{
		err=errors.New("Over Max Retry Still Error!")
	}
	return err
}

func GetAppConfig()*AppConfig  {
	return appConfig
}

func initRefreshInterval() error {
	customizedRefreshInterval:=os.Getenv(refresh_interval_key)
	if isNotEmpty(customizedRefreshInterval){
		interval,err:=strconv.Atoi(customizedRefreshInterval)
		if isNotNil(err) {
			seelog.Errorf("Config for apollo.refreshInterval is invalid:%s",customizedRefreshInterval)
			return err
		}
		refresh_interval=time.Duration(interval)
	}
	return nil
}

func getConfigUrl(config *AppConfig) string{
	return getConfigUrlByHost(config,config.getHost())
}

func getConfigUrlByHost(config *AppConfig,host string) string{
	current:=GetCurrentApolloConfig()
	return fmt.Sprintf("%sconfigs/%s/%s/%s?releaseKey=%s&ip=%s",
		host,
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(config.NamespaceName),
		url.QueryEscape(current.ReleaseKey),
		getInternal())
}

func getNotifyUrl(notifications string,config *AppConfig) string{
	return getNotifyUrlByHost(notifications,
		config,
		config.getHost())
}

func getNotifyUrlByHost(notifications string,config *AppConfig,host string) string{
	return fmt.Sprintf("%snotifications/v2?appId=%s&cluster=%s&notifications=%s",
		host,
		url.QueryEscape(config.AppId),
		url.QueryEscape(config.Cluster),
		url.QueryEscape(notifications))
}

func getServicesConfigUrl(config *AppConfig) string{
	return fmt.Sprintf("%sservices/config?appId=%s&ip=%s",
		config.getHost(),
		url.QueryEscape(config.AppId),
		getInternal())
}