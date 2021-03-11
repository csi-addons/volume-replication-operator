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

package controllers

import (
	"strings"
)

const (
	// Replication Parameters prefixed with replicationParameterPrefix are not passed through
	// to the driver on RPC calls. Instead these are the parameters used by the
	// operator to get the required object from kubernetes and pass it to the
	// Driver.
	replicationParameterPrefix = "replication.storage.openshift.io/"

	prefixedReplicationSecretNameKey      = replicationParameterPrefix + "replication-secret-name"      // name key for secret
	prefixedReplicationSecretNamespaceKey = replicationParameterPrefix + "replication-secret-namespace" // namespace key secret
)

// filterPrefixedParameters removes all the reserved keys from the
// replicationclass which are matching the prefix.
func filterPrefixedParameters(prefix string, param map[string]string) map[string]string {
	newParam := map[string]string{}
	for k, v := range param {
		if !strings.HasPrefix(k, prefix) {
			newParam[k] = v
		}
	}
	return newParam
}
