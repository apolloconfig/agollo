// Copyright 2025 Apollo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/apolloconfig/agollo/v4/component/log"
)

const (
	// fmtInvalidKey is the error message format for invalid key patterns
	fmtInvalidKey = "invalid key format for key %s"
)

var (
	// ErrNilListener represents an error when a nil listener is provided
	ErrNilListener = errors.New("nil listener")
)

// Event represents a configuration change event
// It contains the type of change, the affected key, and the new value
type Event struct {
	EventType ConfigChangeType
	Key       string
	Value     interface{}
}

// Listener is an interface that all event listeners must implement
// It defines the Event method that will be called when configuration changes occur
type Listener interface {
	Event(event *Event)
}

// Dispatcher manages the event distribution system
// It maintains a map of keys to their registered listeners
type Dispatcher struct {
	listeners map[string][]Listener
}

// UseEventDispatch creates and initializes a new event dispatcher
// Returns a new Dispatcher instance with an initialized listeners map
func UseEventDispatch() *Dispatcher {
	eventDispatch := new(Dispatcher)
	eventDispatch.listeners = make(map[string][]Listener)
	return eventDispatch
}

// RegisterListener registers a listener for specific configuration keys
// The keys can be regular expressions to match multiple configuration keys
// Returns an error if the listener is nil or if any key pattern is invalid
func (d *Dispatcher) RegisterListener(listenerObject Listener, keys ...string) error {
	log.Infof("start add  key %v add listener", keys)
	if listenerObject == nil {
		return ErrNilListener
	}

	for _, key := range keys {
		if invalidKey(key) {
			return fmt.Errorf(fmtInvalidKey, key)
		}

		listenerList, ok := d.listeners[key]
		if !ok {
			d.listeners[key] = make([]Listener, 0)
		}

		for _, listener := range listenerList {
			if listener == listenerObject {
				log.Infof("key %s had listener", key)
				return nil
			}
		}
		// append new listener
		listenerList = append(listenerList, listenerObject)
		d.listeners[key] = listenerList
	}
	return nil
}

// invalidKey checks if a key pattern is a valid regular expression
// Returns true if the pattern is invalid, false otherwise
func invalidKey(key string) bool {
	_, err := regexp.Compile(key)
	return err != nil
}

// UnRegisterListener removes a listener from specific configuration keys
// The keys can be regular expressions to match multiple configuration keys
// Returns an error if the listener is nil
func (d *Dispatcher) UnRegisterListener(listenerObj Listener, keys ...string) error {
	if listenerObj == nil {
		return ErrNilListener
	}

	for _, key := range keys {
		listenerList, ok := d.listeners[key]
		if !ok {
			continue
		}

		newListenerList := make([]Listener, 0)
		// remove listener
		for _, listener := range listenerList {
			if listener == listenerObj {
				continue
			}
			newListenerList = append(newListenerList, listener)
		}

		// assign latest listener list
		d.listeners[key] = newListenerList
	}
	return nil
}

// OnChange implements the ChangeEvent handler for Apollo configuration changes
// It processes the change event and dispatches events to registered listeners
func (d *Dispatcher) OnChange(changeEvent *ChangeEvent) {
	if changeEvent == nil {
		return
	}
	log.Logger.Infof("get change event for namespace %s", changeEvent.Namespace)
	for key, event := range changeEvent.Changes {
		d.dispatchEvent(key, event)
	}
}

// OnNewestChange handles the latest configuration change events
// This method is currently empty and reserved for future implementation
func (d *Dispatcher) OnNewestChange(event *FullChangeEvent) {

}

// dispatchEvent dispatches a configuration change event to all matching listeners
// It matches the event key against registered key patterns and notifies matching listeners
func (d *Dispatcher) dispatchEvent(eventKey string, event *ConfigChange) {
	for regKey, listenerList := range d.listeners {
		matched, err := regexp.MatchString(regKey, eventKey)
		if err != nil {
			log.Logger.Errorf("regular expression for key %s, error: %v", eventKey, err)
			continue
		}
		if matched {
			for _, listener := range listenerList {
				log.Logger.Infof("event generated for %s key %s", regKey, eventKey)
				go listener.Event(convertToEvent(eventKey, event))
			}
		}
	}
}

// convertToEvent converts a ConfigChange to an Event
// It sets the appropriate value based on the change type
func convertToEvent(key string, event *ConfigChange) *Event {
	e := &Event{
		EventType: event.ChangeType,
		Key:       key,
	}
	switch event.ChangeType {
	case ADDED:
		e.Value = event.NewValue
	case MODIFIED:
		e.Value = event.NewValue
	case DELETED:
		e.Value = event.OldValue
	}
	return e
}
