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

// DisableTask stores configuration required to disable volume replication
type DisableTask struct {
	// params required for this task
	CommonRequestParameters
}

// NewDisableTask returns a DisableTask object
func NewDisableTask(c CommonRequestParameters) *DisableTask {
	return &DisableTask{c}
}

// Run is used to run Disable task
func (d *DisableTask) Run() error {
	_, err := d.Replication.DisableVolumeReplication(
		d.CommonRequestParameters.VolumeID,
		d.CommonRequestParameters.Secrets,
		d.CommonRequestParameters.Parameters,
	)
	return err
}
