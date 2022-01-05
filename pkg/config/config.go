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

package config

import (
	"errors"
	"time"
)

// DriverConfig holds the CSI driver configuration.
type DriverConfig struct {
	// DriverEndpoint represents the gRPC endpoint for CSI driver.
	DriverEndpoint string
	// DriverName is the name of CSI Driver.
	DriverName string
	// RPCTimeout for RPCs to the CSI driver.
	RPCTimeout time.Duration
}

// NewDriverConfig returns the newly initialized DriverConfig.
func NewDriverConfig() *DriverConfig {
	return &DriverConfig{}
}

// Validate operation configurations.
func (cfg *DriverConfig) Validate() error {
	// check driver name is set
	if cfg.DriverName == "" {
		return errors.New("driverName is empty")
	}

	return nil
}
