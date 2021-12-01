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
	"github.com/csi-addons/volume-replication-operator/pkg/client"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Replication represents the instance of a single replication operation
type Replication struct {
	Params CommonRequestParameters
	Force  bool
}

// ReplicationResponse is the response of a replication operation
type ReplicationResponse struct {
	Response interface{}
	Error    error
}

// CommonRequestParameters holds the common parameters across replication operations.
type CommonRequestParameters struct {
	VolumeID      string
	ReplicationID string
	Parameters    map[string]string
	Secrets       map[string]string
	Replication   client.VolumeReplication
}

func (r *Replication) Enable() *ReplicationResponse {
	resp, err := r.Params.Replication.EnableVolumeReplication(
		r.Params.VolumeID,
		r.Params.ReplicationID,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &ReplicationResponse{Response: resp, Error: err}
}

func (r *Replication) Disable() *ReplicationResponse {
	resp, err := r.Params.Replication.DisableVolumeReplication(
		r.Params.VolumeID,
		r.Params.ReplicationID,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &ReplicationResponse{Response: resp, Error: err}
}

func (r *Replication) Promote() *ReplicationResponse {
	resp, err := r.Params.Replication.PromoteVolume(
		r.Params.VolumeID,
		r.Params.ReplicationID,
		r.Force,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &ReplicationResponse{Response: resp, Error: err}
}

func (r *Replication) Demote() *ReplicationResponse {
	resp, err := r.Params.Replication.DemoteVolume(
		r.Params.VolumeID,
		r.Params.ReplicationID,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &ReplicationResponse{Response: resp, Error: err}
}

func (r *Replication) Resync() *ReplicationResponse {
	resp, err := r.Params.Replication.ResyncVolume(
		r.Params.VolumeID,
		r.Params.ReplicationID,
		r.Params.Secrets,
		r.Params.Parameters,
	)

	return &ReplicationResponse{Response: resp, Error: err}
}

func (r *ReplicationResponse) HasKnownGRPCError(knownErrors []codes.Code) bool {
	if r.Error == nil {
		return false
	}

	s, ok := status.FromError(r.Error)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		return false
	}

	for _, e := range knownErrors {
		if s.Code() == e {
			return true
		}
	}

	return false
}

// GetMessageFromError returns the message from the error.
func GetMessageFromError(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		return err.Error()
	}

	return s.Message()
}
