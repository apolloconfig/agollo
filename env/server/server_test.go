/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"testing"
	"time"

	. "github.com/tevid/gohamcrest"

	"github.com/apolloconfig/agollo/v4/env/config"
)

var name = "abc"

func TestGetServersLen(t *testing.T) {
	m := make(map[string]*config.ServerInfo, 2)
	m["b"] = &config.ServerInfo{}
	m["c"] = &config.ServerInfo{}
	SetServers(name, m)
	serversLen := GetServersLen(name)
	Assert(t, serversLen, Equal(2))
}

func TestSetNextTryConnTime(t *testing.T) {
	SetNextTryConnTime(name, 10)
	Assert(t, int(ipMap[name].nextTryConnTime), GreaterThan(int(time.Now().Unix())))
}

func TestAppConfig_IsConnectDirectly(t *testing.T) {
	s := &Info{
		serverMap:       nil,
		nextTryConnTime: 0,
	}
	ipMap[name] = s
	isConnectDirectly := IsConnectDirectly(name)
	Assert(t, isConnectDirectly, Equal(false))

	SetNextTryConnTime(name, 10)
	isConnectDirectly = IsConnectDirectly(name)
	Assert(t, isConnectDirectly, Equal(false))
}
