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
	"time"
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

func createChangeEvent() *ChangeEvent {
	addConfig := createAddConfigChange("new")
	deleteConfig := createDeletedConfigChange("old")
	modifyConfig := createModifyConfigChange("old", "new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	changes["adx"] = addConfig
	changes["delete"] = deleteConfig
	changes["modify"] = modifyConfig
	cEvent := &ChangeEvent{
		Changes: changes,
	}
	cEvent.Namespace = "a"
	return cEvent
}

func TestUseDispatch(t *testing.T) {
	UseEventDispatch()
	Assert(t, changeListeners.Len(), Equal(1))
	RemoveChangeListener(eventDispatch)
}

func TestDispatch(t *testing.T) {
	UseEventDispatch()
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err := RegisterListener(l, "add", "delete")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(2))
	cEvent := createChangeEvent()
	pushChangeEvent(cEvent)
	time.Sleep(1 * time.Second)
	Assert(t, len(l.Keys), Equal(2))
	v, ok := l.Keys["add"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
	v, ok = l.Keys["delete"]
	Assert(t, ok, Equal(true))
	Assert(t, v, Equal("old"))
	_, ok = l.Keys["modify"]
	Assert(t, ok, Equal(false))
}

func TestRegDispatch(t *testing.T) {
	UseEventDispatch()
	err := RegisterListener(nil, "ad.*")
	Assert(t, err, NotNilVal())
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err = RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(1))
	cEvent := createChangeEvent()
	pushChangeEvent(cEvent)
	time.Sleep(1 * time.Second)
	Assert(t, len(l.Keys), Equal(2))
	v, ok := l.Keys["add"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
	v, ok = l.Keys["adx"]
	Assert(t, v, Equal("new"))
	Assert(t, ok, Equal(true))
}

func TestDuplicateRegDispatch(t *testing.T) {
	UseEventDispatch()
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err := RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(1))
	Assert(t, len(eventDispatch.listeners["ad.*"]), Equal(1))

	err = RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(1))
	Assert(t, len(eventDispatch.listeners["ad.*"]), Equal(1))
}

func TestUnRegisterListener(t *testing.T) {
	UseEventDispatch()
	err := RegisterListener(nil, "ad.*")
	Assert(t, err, NotNilVal())
	l := &CustomListener{
		Keys: make(map[string]interface{}, 0),
	}
	err = RegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(1))
	Assert(t, len(eventDispatch.listeners["ad.*"]), Equal(1))

	err = UnRegisterListener(l, "ad.*")
	Assert(t, err, NilVal())
	Assert(t, len(eventDispatch.listeners), Equal(1))
	Assert(t, len(eventDispatch.listeners["ad.*"]), Equal(0))

}
