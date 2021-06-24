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

package v1alpha1

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var volumereplicationclasslog = logf.Log.WithName("volumereplicationclass-resource")

func (r *VolumeReplicationClass) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-replication-storage-openshift-io-v1alpha1-volumereplicationclass,mutating=true,failurePolicy=fail,sideEffects=None,groups=replication.storage.openshift.io,resources=volumereplicationclasses,verbs=create;update,versions=v1alpha1,name=mvolumereplicationclass.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &VolumeReplicationClass{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *VolumeReplicationClass) Default() {
	volumereplicationclasslog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-replication-storage-openshift-io-v1alpha1-volumereplicationclass,mutating=false,failurePolicy=fail,sideEffects=None,groups=replication.storage.openshift.io,resources=volumereplicationclasses,verbs=create;update,versions=v1alpha1,name=vvolumereplicationclass.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &VolumeReplicationClass{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *VolumeReplicationClass) ValidateCreate() error {
	volumereplicationclasslog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *VolumeReplicationClass) ValidateUpdate(old runtime.Object) error {
	volumereplicationclasslog.Info("validate update", "name", r.Name)

	oldObj := old.(*VolumeReplicationClass)

	if r.Spec.Provisioner != oldObj.Spec.Provisioner {
		err := fmt.Errorf("Spec.Provisioner is immutable")
		volumereplicationclasslog.Error(err, "Spec.Provisioner can not be changed")
		return err
	}

	if !reflect.DeepEqual(r.Spec.Parameters, oldObj.Spec.Parameters) {
		err := fmt.Errorf("Spec.Parameters is immutable")
		volumereplicationclasslog.Error(err, "Spec.Parameters can not be changed")
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *VolumeReplicationClass) ValidateDelete() error {
	volumereplicationclasslog.Info("validate delete", "name", r.Name)

	return nil
}
