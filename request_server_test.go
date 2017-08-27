package agollo

import (
	"net/http"
	"fmt"
	"encoding/json"
)

const servicesResponseStr  =`[{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.128.102:apollo-configservice:8080",
"homepageUrl": "http://localhost:7080/"
},
{
"appName": "APOLLO-CONFIGSERVICE",
"instanceId": "10.15.88.125:apollo-configservice:8080",
"homepageUrl": "http://localhost:7081/"
}]`

//run mock config server
func runMockServicesServer(handler func(http.ResponseWriter, *http.Request)) {
	tmpServerInfo:=make([]*serverInfo,0)

	json.Unmarshal([]byte(servicesResponseStr),&tmpServerInfo)

	for _,server := range tmpServerInfo{
		uri:=fmt.Sprintf("/services/config")
		http.HandleFunc(uri, handler)

		http.ListenAndServe(fmt.Sprintf("%s",server.HomepageUrl), nil)
	}
}

func closeMockServicesServer() {
	http.DefaultServeMux=&http.ServeMux{}
}


//Normal response
func normalServicesResponse(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, servicesResponseStr)
}