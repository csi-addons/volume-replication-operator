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
	"errors"
	"testing"
	"time"

	replicationlib "github.com/csi-addons/spec/lib/go/replication"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/csi-addons/volume-replication-operator/pkg/client/fake"
)

func TestEnableVolumeReplication(t *testing.T) {
	var client = NewReplicationClient(&grpc.ClientConn{}, time.Minute)
	// return success response
	mockedEnableReplication := &fake.ReplicationClient{
		EnableVolumeReplicationMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.EnableVolumeReplicationResponse, error) {
			return &replicationlib.EnableVolumeReplicationResponse{}, nil
		},
	}
	client = mockedEnableReplication
	resp, err := client.EnableVolumeReplication("", "", nil, nil)
	assert.Equal(t, &replicationlib.EnableVolumeReplicationResponse{}, resp)
	assert.Nil(t, err)

	// return error
	mockedEnableReplication = &fake.ReplicationClient{
		EnableVolumeReplicationMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.EnableVolumeReplicationResponse, error) {
			return nil, errors.New("failed to enable mirroring")
		},
	}
	client = mockedEnableReplication
	resp, err = client.EnableVolumeReplication("", "", nil, nil)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func TestDisableVolumeReplication(t *testing.T) {
	var client = NewReplicationClient(&grpc.ClientConn{}, time.Minute)
	// return success response
	mockedDisableReplication := &fake.ReplicationClient{
		DisableVolumeReplicationMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.DisableVolumeReplicationResponse, error) {
			return &replicationlib.DisableVolumeReplicationResponse{}, nil
		},
	}
	client = mockedDisableReplication
	resp, err := client.DisableVolumeReplication("", "", nil, nil)
	assert.Equal(t, &replicationlib.DisableVolumeReplicationResponse{}, resp)
	assert.Nil(t, err)

	// return error
	mockedDisableReplication = &fake.ReplicationClient{
		DisableVolumeReplicationMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.DisableVolumeReplicationResponse, error) {
			return nil, errors.New("failed to disable mirroring")
		},
	}
	client = mockedDisableReplication
	resp, err = client.DisableVolumeReplication("", "", nil, nil)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func TestPromoteVolume(t *testing.T) {
	var client = NewReplicationClient(&grpc.ClientConn{}, time.Minute)
	// return success response
	mockedPromoteVolume := &fake.ReplicationClient{
		PromoteVolumeMock: func(volumeID, replicationID string, force bool, secrets, parameters map[string]string) (*replicationlib.PromoteVolumeResponse, error) {
			return &replicationlib.PromoteVolumeResponse{}, nil
		},
	}
	force := false
	client = mockedPromoteVolume
	resp, err := client.PromoteVolume("", "", force, nil, nil)
	assert.Equal(t, &replicationlib.PromoteVolumeResponse{}, resp)
	assert.Nil(t, err)

	// return error
	mockedPromoteVolume = &fake.ReplicationClient{
		PromoteVolumeMock: func(volumeID, replicationID string, force bool, secrets, parameters map[string]string) (*replicationlib.PromoteVolumeResponse, error) {
			return nil, errors.New("failed to promote volume")
		},
	}
	client = mockedPromoteVolume
	resp, err = client.PromoteVolume("", "", force, nil, nil)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func TestDemoteVolume(t *testing.T) {
	var client = NewReplicationClient(&grpc.ClientConn{}, time.Minute)
	// return success response
	mockedDemoteVolume := &fake.ReplicationClient{
		DemoteVolumeMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.DemoteVolumeResponse, error) {
			return &replicationlib.DemoteVolumeResponse{}, nil
		},
	}
	client = mockedDemoteVolume
	resp, err := client.DemoteVolume("", "", nil, nil)
	assert.Equal(t, &replicationlib.DemoteVolumeResponse{}, resp)
	assert.Nil(t, err)

	// return error
	mockedDemoteVolume = &fake.ReplicationClient{
		DemoteVolumeMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.DemoteVolumeResponse, error) {
			return nil, errors.New("failed to demote volume")
		},
	}
	client = mockedDemoteVolume
	resp, err = client.DemoteVolume("", "", nil, nil)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}

func TestResyncVolume(t *testing.T) {
	var client = NewReplicationClient(&grpc.ClientConn{}, time.Minute)
	// return success response
	mockedResyncVolume := &fake.ReplicationClient{
		ResyncVolumeMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.ResyncVolumeResponse, error) {
			return &replicationlib.ResyncVolumeResponse{}, nil
		},
	}
	client = mockedResyncVolume
	resp, err := client.ResyncVolume("", "", nil, nil)
	assert.Equal(t, &replicationlib.ResyncVolumeResponse{}, resp)
	assert.Nil(t, err)

	// return error
	mockedResyncVolume = &fake.ReplicationClient{
		ResyncVolumeMock: func(volumeID, replicationID string, secrets, parameters map[string]string) (*replicationlib.ResyncVolumeResponse, error) {
			return nil, errors.New("failed to resync volume")
		},
	}
	client = mockedResyncVolume
	resp, err = client.ResyncVolume("", "", nil, nil)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
}
