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

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetMessageFromError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "test GRPC error message",
			err:  status.Error(codes.Internal, "failure"),
			want: "failure",
		},
		{
			name: "test non grpc error message",
			err:  errors.New("non grpc failure"),
			want: "non grpc failure",
		},
		{
			name: "test nil error",
			err:  nil,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMessageFromError(tt.err); got != tt.want {
				t.Errorf("GetMessageFromError() = %v, want %v", got, tt.want)
			}
		})
	}
}
