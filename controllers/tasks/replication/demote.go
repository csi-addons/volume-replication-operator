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

// DemoteVolumeTask stores configuration required to demote volume
type DemoteVolumeTask struct {
	// params required for this task
	CommonRequestParameters
}

// NewDemoteVolumeTask returns a DemoteVolumeTask object
func NewDemoteVolumeTask(c CommonRequestParameters) *DemoteVolumeTask {
	return &DemoteVolumeTask{c}
}

// Run is used to run DemoteVolume task
func (d *DemoteVolumeTask) Run() (interface{}, error) {
	resp, err := d.Replication.DemoteVolume(
		d.CommonRequestParameters.VolumeID,
		d.CommonRequestParameters.Secrets,
		d.CommonRequestParameters.Parameters,
	)
	return resp, err
}
