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
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// getSecret retrives the secrets based on name and namespace input
func (r *VolumeReplicationReconciler) getSecret(name, namespace string) (map[string][]byte, error) {
	namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
	secret := &corev1.Secret{}
	err := r.Client.Get(context.TODO(), namespacedName, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Error(err, "Secret not found", "Secret Name", name, "Secret Namespace", namespace)
		}
		r.Log.Error(err, "error getting secret", "Secret Name", name, "Secret Namespace", namespace)
		return nil, err
	}
	return secret.Data, nil
}
