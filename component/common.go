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

package component

// AbsComponent defines the interface for components that require periodic execution
// This interface is used by various Apollo components that need to run continuously
// or on a scheduled basis, such as configuration synchronization and server list updates
type AbsComponent interface {
	// Start initiates the component's main execution loop
	// Implementations should handle their own scheduling and error recovery
	Start()
}

// StartRefreshConfig begins the execution of a periodic component
// Parameters:
//   - component: An implementation of AbsComponent that needs to be started
//
// This function is responsible for initiating the component's execution cycle
// and is typically called during system initialization
func StartRefreshConfig(component AbsComponent) {
	component.Start()
}
