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

package storage

import (
	. "github.com/tevid/gohamcrest"
	"sync"
	"testing"
)

type CustomListener struct {
	l    sync.Mutex
	Keys map[string]interface{}
}

func (t *CustomListener) Event(event *Event) {
	t.l.Lock()
	defer t.l.Unlock()
	t.Keys[event.Key] = event.Value
}

func TestDispatch(t *testing.T) {
	dispatch := UseEventDispatch()
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err := dispatch.RegisterListener(l, "add", "delete")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(2))
}

func TestRegDispatch(t *testing.T) {
	dispatch := UseEventDispatch()
	err := dispatch.RegisterListener(nil, "ad.*")
	Assert(t, err, NotNilVal())
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err = dispatch.RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(1))
}

func TestDuplicateRegDispatch(t *testing.T) {
	dispatch := UseEventDispatch()
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err := dispatch.RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(1))
	Assert(t, len(dispatch.listeners["ad.*"]), Equal(1))

	err = dispatch.RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(1))
	Assert(t, len(dispatch.listeners["ad.*"]), Equal(1))
}

func TestUnRegisterListener(t *testing.T) {
	dispatch := UseEventDispatch()
	err := dispatch.RegisterListener(nil, "ad.*")
	Assert(t, err, NotNilVal())
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err = dispatch.RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(1))
	Assert(t, len(dispatch.listeners["ad.*"]), Equal(1))

	err = dispatch.UnRegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(dispatch.listeners), Equal(1))
	Assert(t, len(dispatch.listeners["ad.*"]), Equal(0))

}
