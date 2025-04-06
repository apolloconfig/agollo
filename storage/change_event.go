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

const (
	// ADDED represents a new configuration item being added
	ADDED ConfigChangeType = iota
	// MODIFIED represents an existing configuration item being modified
	MODIFIED
	// DELETED represents an existing configuration item being removed
	DELETED
)

// ChangeListener defines the interface for configuration change event handlers
// Implementations can react to both incremental and full configuration changes
type ChangeListener interface {
	// OnChange handles incremental configuration changes
	// Parameters:
	//   - event: Contains details about specific property changes
	OnChange(event *ChangeEvent)

	// OnNewestChange handles full configuration changes
	// Parameters:
	//   - event: Contains the complete new configuration state
	OnNewestChange(event *FullChangeEvent)
}

// ConfigChangeType represents the type of configuration change
type ConfigChangeType int

// baseChangeEvent contains common fields for all change events
type baseChangeEvent struct {
	// Namespace identifies the configuration namespace
	Namespace string
	// NotificationID is the unique identifier for this change notification
	NotificationID int64
}

// ChangeEvent represents an incremental configuration change
// It contains information about specific property changes
type ChangeEvent struct {
	baseChangeEvent
	// Changes maps property keys to their change details
	Changes map[string]*ConfigChange
}

// ConfigChange contains details about a single property's change
type ConfigChange struct {
	// OldValue contains the previous value of the property
	OldValue interface{}
	// NewValue contains the updated value of the property
	NewValue interface{}
	// ChangeType indicates how the property changed (ADDED/MODIFIED/DELETED)
	ChangeType ConfigChangeType
}

// FullChangeEvent represents a complete configuration state change
type FullChangeEvent struct {
	baseChangeEvent
	// Changes maps property keys to their new values
	Changes map[string]interface{}
}

// createModifyConfigChange creates a ConfigChange for modified properties
// Parameters:
//   - oldValue: The previous value of the property
//   - newValue: The updated value of the property
//
// Returns:
//   - *ConfigChange: A change record with MODIFIED type
func createModifyConfigChange(oldValue interface{}, newValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeType: MODIFIED,
	}
}

// createAddConfigChange creates a ConfigChange for new properties
// Parameters:
//   - newValue: The value of the new property
//
// Returns:
//   - *ConfigChange: A change record with ADDED type
func createAddConfigChange(newValue interface{}) *ConfigChange {
	return &ConfigChange{
		NewValue:   newValue,
		ChangeType: ADDED,
	}
}

// createDeletedConfigChange creates a ConfigChange for removed properties
// Parameters:
//   - oldValue: The last value of the property before deletion
//
// Returns:
//   - *ConfigChange: A change record with DELETED type
func createDeletedConfigChange(oldValue interface{}) *ConfigChange {
	return &ConfigChange{
		OldValue:   oldValue,
		ChangeType: DELETED,
	}
}

// createConfigChangeEvent creates a new ChangeEvent from a set of changes
// Parameters:
//   - changes: Map of property keys to their change details
//   - nameSpace: The configuration namespace
//   - notificationID: Unique identifier for this change notification
//
// Returns:
//   - *ChangeEvent: A new change event containing all changes
func createConfigChangeEvent(changes map[string]*ConfigChange, nameSpace string, notificationID int64) *ChangeEvent {
	c := &ChangeEvent{
		Changes: changes,
	}
	c.Namespace = nameSpace
	c.NotificationID = notificationID
	return c
}
