package agollo

import (
	"net/http"
	"fmt"
	//"time"
)

const servicesConfigResponseStr  =`[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.102:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.125:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.11:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.11:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.193:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.193:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.101:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.101:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.192:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.192:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.124:apollo-configservice:8080",
"homepageUrl": "http://10.15.88.124:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.103:apollo-configservice:8080",
"homepageUrl": "http://10.15.128.103:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "localhost:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.12:8080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.14.0.194:apollo-configservice:8080",
"homepageUrl": "http://10.14.0.194:8080/"
}
]`

//var server *http.Server

//run mock config server
func runMockServicesConfigServer(handler func(http.ResponseWriter, *http.Request)) {
	uri:=fmt.Sprintf("/services/config")
	http.HandleFunc(uri, handler)

	//server = &http.Server{
	//	Addr:    appConfig.Ip,
	//	Handler: http.DefaultServeMux,
	//}
	//
	//server.ListenAndServe()


	logger.Info("mock notify server:",appConfig.Ip)
	err:=http.ListenAndServe(fmt.Sprintf("%s",appConfig.Ip), nil)
	if err!=nil{
		logger.Error("runMockConfigServer err:",err)
	}
}

func closeMockServicesConfigServer() {
	http.DefaultServeMux=http.NewServeMux()
	//server.Close()
}


//Normal response
func normalServicesConfigResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, servicesConfigResponseStr)
}

////Error response
////will hold 5s and keep response 404
//func errorConfigResponse(rw http.ResponseWriter, req *http.Request) {
//	time.Sleep(500 * time.Microsecond)
//	rw.WriteHeader(http.StatusNotFound)
//}