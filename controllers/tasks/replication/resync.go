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

// ResyncVolumeTask stores configuration required to resync volume
type ResyncVolumeTask struct {
	// params required for this task
	CommonRequestParameters
}

// NewResyncVolumeTask returns a ResyncVolumeTask object
func NewResyncVolumeTask(c CommonRequestParameters) *ResyncVolumeTask {
	return &ResyncVolumeTask{c}
}

// Run is used to run ResyncVolume task
func (r *ResyncVolumeTask) Run() error {
	_, err := r.Replication.ResyncVolume(
		r.CommonRequestParameters.VolumeID,
		r.CommonRequestParameters.Secrets,
		r.CommonRequestParameters.Parameters,
	)
	return err
}
