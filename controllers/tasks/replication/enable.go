/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package replication

// EnableTask stores configuration required to enable volume replication
type EnableTask struct {
	// params required for this task
	CommonRequestParameters
}

// NewEnableTask returns a EnableTask object
func NewEnableTask(c CommonRequestParameters) *EnableTask {
	return &EnableTask{c}
}

// Run is used to run Enable task
func (e *EnableTask) Run() (interface{}, error) {
	// perform sub-tasks
	resp, err := e.Replication.EnableVolumeReplication(
		e.CommonRequestParameters.VolumeID,
		e.CommonRequestParameters.Secrets,
		e.CommonRequestParameters.Parameters,
	)

	return resp, err
}
