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

package client

import (
	"context"
	"time"

	replicationlib "github.com/csi-addons/spec/lib/go/replication"
	"google.golang.org/grpc"
)

type replicationClient struct {
	client  replicationlib.ControllerClient
	timeout time.Duration
}

// VolumeReplication holds the methods required for volume replication.
type VolumeReplication interface {
	// EnableVolumeReplication RPC call to enable the volume replication.
	EnableVolumeReplication(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.EnableVolumeReplicationResponse, error)
	// DisableVolumeReplication RPC call to disable the volume replication.
	DisableVolumeReplication(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.DisableVolumeReplicationResponse, error)
	// PromoteVolume RPC call to promote the volume.
	PromoteVolume(volumeID, replicationID string, force bool, secrets, parameters map[string]string) (*replicationlib.
		PromoteVolumeResponse, error)
	// DemoteVolume RPC call to demote the volume.
	DemoteVolume(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.
		DemoteVolumeResponse, error)
	// ResyncVolume RPC call to resync the volume.
	ResyncVolume(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.
		ResyncVolumeResponse, error)
}

// NewReplicationClient returns VolumeReplication interface which has the RPC
// calls for replication.
func NewReplicationClient(cc *grpc.ClientConn, timeout time.Duration) VolumeReplication {
	return &replicationClient{client: replicationlib.NewControllerClient(cc), timeout: timeout}
}

// EnableVolumeReplication RPC call to enable the volume replication.
func (rc *replicationClient) EnableVolumeReplication(volumeID, replicationID string,
	secrets, parameters map[string]string) (*replicationlib.EnableVolumeReplicationResponse, error) {
	req := &replicationlib.EnableVolumeReplicationRequest{
		VolumeId:      volumeID,
		ReplicationId: replicationID,
		Parameters:    parameters,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.EnableVolumeReplication(createCtx, req)

	return resp, err
}

// DisableVolumeReplication RPC call to disable the volume replication.
func (rc *replicationClient) DisableVolumeReplication(volumeID, replicationID string,
	secrets, parameters map[string]string) (*replicationlib.DisableVolumeReplicationResponse, error) {
	req := &replicationlib.DisableVolumeReplicationRequest{
		VolumeId:      volumeID,
		ReplicationId: replicationID,
		Parameters:    parameters,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.DisableVolumeReplication(createCtx, req)

	return resp, err
}

// PromoteVolume RPC call to promote the volume.
func (rc *replicationClient) PromoteVolume(volumeID, replicationID string,
	force bool, secrets, parameters map[string]string) (*replicationlib.PromoteVolumeResponse, error) {
	req := &replicationlib.PromoteVolumeRequest{
		VolumeId:      volumeID,
		ReplicationId: replicationID,
		Force:         force,
		Parameters:    parameters,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.PromoteVolume(createCtx, req)

	return resp, err
}

// DemoteVolume RPC call to demote the volume.
func (rc *replicationClient) DemoteVolume(volumeID, replicationID string,
	secrets, parameters map[string]string) (*replicationlib.DemoteVolumeResponse, error) {
	req := &replicationlib.DemoteVolumeRequest{
		VolumeId:      volumeID,
		ReplicationId: replicationID,
		Parameters:    parameters,
		Secrets:       secrets,
	}
	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.DemoteVolume(createCtx, req)

	return resp, err
}

// ResyncVolume RPC call to resync the volume.
func (rc *replicationClient) ResyncVolume(volumeID, replicationID string,
	secrets, parameters map[string]string) (*replicationlib.ResyncVolumeResponse, error) {
	req := &replicationlib.ResyncVolumeRequest{
		VolumeId:      volumeID,
		ReplicationId: replicationID,
		Parameters:    parameters,
		Secrets:       secrets,
	}

	createCtx, cancel := context.WithTimeout(context.Background(), rc.timeout)
	defer cancel()
	resp, err := rc.client.ResyncVolume(createCtx, req)

	return resp, err
}
