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

package tasks

// TaskSpec is the specification for each Task
type TaskSpec struct {
	Name        string
	Task        Task
	KnownErrors []error
}

// Task is a specific task to be done by controller
type Task interface {
	Run() error
}

// RunAll executes all the Task in the given list of TaskSpec
func RunAll(tasks []*TaskSpec) (string, error) {
	for _, task := range tasks {
		if err := task.Task.Run(); err != nil {
			// if err is in KnownErrors then continue
			// else return
			return task.Name, err
		}
	}
	return "", nil
}
