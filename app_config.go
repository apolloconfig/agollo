package agollo

import (
	"os"
	"strconv"
	"time"
	"fmt"
	"net/url"
	"encoding/json"
)

const appConfigFileName  ="app.properties"

var (
	refresh_interval = 5 *time.Minute //5m
	refresh_interval_key = "apollo.refreshInterval"  //

	long_poll_interval = 2 *time.Second //2s
	long_poll_connect_timeout  = 1 * time.Minute //1m

	connect_timeout  = 1 * time.Second //1s
	//notify timeout
	nofity_connect_timeout  = 10 * time.Minute //10m
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

	//load app config method
	loadAppConfig func()(*AppConfig,error)
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
func (this *AppConfig) setNextTryConnTime(nextTryConnectPeriod int64){
	this.NextTryConnTime=time.Now().Unix()+nextTryConnectPeriod
}

//is connect by ip directly
//false : no
//true : yes
func (this *AppConfig) isConnectDirectly() bool{
	if this.NextTryConnTime>=0&&this.NextTryConnTime>time.Now().Unix(){
		return true
	}

	return false
}

func (this *AppConfig) selectHost() string{
	if !this.isConnectDirectly(){
		return this.getHost()
	}

	for host,server:=range servers{
		// if some node has down then select next node
		if server.IsDown{
			continue
		}
		return host
	}

	return ""
}

func setDownNode(host string) {
	if host=="" || appConfig==nil{
		return
	}

	if host==appConfig.getHost(){
		appConfig.setNextTryConnTime(next_try_connect_period)
	}

	for key,server:=range servers{
		if key==host{
			server.IsDown=true
			break
		}
	}
}


type serverInfo struct {
	AppName string `json:"appName"`
	InstanceId string `json:"instanceId"`
	HomepageUrl string `json:"homepageUrl"`
	IsDown bool `json:"-"`

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
	appConfig,err = getLoadAppConfig()

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

// set load app config's function
func getLoadAppConfig() (*AppConfig,error) {
	if loadAppConfig!=nil{
		return loadAppConfig()
	}
	return loadJsonConfig(appConfigFileName)
}

// set load app config's function
func SetLoadAppConfig(load func()(*AppConfig,error)) {
	loadAppConfig=load
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

func syncServerIpListSuccessCallBack(responseBody []byte)(o interface{},err error){
	tmpServerInfo:=make([]*serverInfo,0)

	err= json.Unmarshal(responseBody,&tmpServerInfo)

	if err!=nil{
		logger.Error("Unmarshal json Fail,Error:",err)
		return
	}

	if len(tmpServerInfo)==0 {
		logger.Info("get no real server!")
		return
	}

	for _,server :=range tmpServerInfo {
		if server==nil{
			continue
		}
		servers[server.HomepageUrl]=server
	}
	return
}

//sync ip list from server
//then
//1.update cache
//2.store in disk
func syncServerIpList() error{
	appConfig:=GetAppConfig()
	if appConfig==nil{
		panic("can not find apollo config!please confirm!")
	}
	url:=getServicesConfigUrl(appConfig)
	logger.Debug("url:",url)

	_,err:=request(url,&ConnectConfig{
	},&CallBack{
		SuccessCallBack:syncServerIpListSuccessCallBack,
	})


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
			logger.Errorf("Config for apollo.refreshInterval is invalid:%s",customizedRefreshInterval)
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

func getConfigUrlSuffix(config *AppConfig) string{
	current:=GetCurrentApolloConfig()
	return fmt.Sprintf("configs/%s/%s/%s?releaseKey=%s&ip=%s",
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

func getNotifyUrlSuffix(notifications string,config *AppConfig) string{
	return fmt.Sprintf("notifications/v2?appId=%s&cluster=%s&notifications=%s",
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