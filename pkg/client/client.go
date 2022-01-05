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

	"github.com/kubernetes-csi/csi-lib-utils/connection"
	"github.com/kubernetes-csi/csi-lib-utils/metrics"
	"github.com/kubernetes-csi/csi-lib-utils/rpc"
	"google.golang.org/grpc"
)

// Client holds the GRPC connenction details.
type Client struct {
	Client  *grpc.ClientConn
	Timeout time.Duration
}

// Connect to the GRPC client.
func connect(address string) (*grpc.ClientConn, error) {
	return connection.Connect(address, metrics.NewCSIMetricsManager(""), connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
}

// New creates and returns the GRPC client.
func New(address string, timeout time.Duration) (*Client, error) {
	c := &Client{}
	cc, err := connect(address)
	if err != nil {
		return c, err
	}
	c.Client = cc
	c.Timeout = timeout

	return c, nil
}

// Probe the GRPC client once.
func (c *Client) Probe() error {
	return rpc.ProbeForever(c.Client, c.Timeout)
}

// GetDriverName gets the driver name from the driver.
func (c *Client) GetDriverName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return rpc.GetDriverName(ctx, c.Client)
}
