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
	"sync"
	"testing"

	. "github.com/tevid/gohamcrest"
)

var listener *CustomChangeListener

func init() {
	listener = &CustomChangeListener{}
}

type CustomChangeListener struct {
	w sync.WaitGroup
}

func (t *CustomChangeListener) OnChange(event *ChangeEvent) {
	t.w.Done()
}

func (t *CustomChangeListener) OnNewestChange(event *FullChangeEvent) {

}

func TestAddChangeListener(t *testing.T) {

	AddChangeListener(nil)
	Assert(t, changeListeners.Len(), Equal(0))

	AddChangeListener(listener)

	Assert(t, changeListeners.Len(), Equal(1))
}

func TestGetChangeListeners(t *testing.T) {
	Assert(t, GetChangeListeners().Len(), Equal(1))
}

func TestRemoveChangeListener(t *testing.T) {
	RemoveChangeListener(nil)
	Assert(t, changeListeners.Len(), Equal(1))
	RemoveChangeListener(listener)
	Assert(t, changeListeners.Len(), Equal(0))
}

func TestPushChangeEvent(t *testing.T) {

	addConfig := createAddConfigChange("new")
	deleteConfig := createDeletedConfigChange("old")
	modifyConfig := createModifyConfigChange("old", "new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	event := &ChangeEvent{
		Changes: changes,
	}
	event.Namespace = "a"
	changes["delete"] = deleteConfig
	changes["modify"] = modifyConfig
	listener = &CustomChangeListener{}
	listener.w.Add(1)

	AddChangeListener(listener)

	pushChangeEvent(event)

	listener.w.Wait()

	RemoveChangeListener(listener)
}

func TestCreateConfigChangeEvent(t *testing.T) {
	addConfig := createAddConfigChange("new")
	changes := make(map[string]*ConfigChange)
	changes["add"] = addConfig
	event := createConfigChangeEvent(changes, "ns")
	Assert(t, event, NotNilVal())
	Assert(t, len(event.Changes), Equal(1))
	Assert(t, event.Namespace, Equal("ns"))
}
