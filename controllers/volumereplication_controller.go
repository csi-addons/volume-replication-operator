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
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	replicationv1alpha1 "github.com/kube-storage/volume-replication-operator/api/v1alpha1"
	"github.com/kube-storage/volume-replication-operator/controllers/tasks"
	"github.com/kube-storage/volume-replication-operator/controllers/tasks/replication"
	"github.com/kube-storage/volume-replication-operator/pkg/config"
)

const (
	pvcDataSource = "PersistentVolumeClaim"
)

// VolumeReplicationReconciler reconciles a VolumeReplication object
type VolumeReplicationReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DriverConfig *config.DriverConfig
}

// +kubebuilder:rbac:groups=replication.storage.openshift.io,resources=volumereplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=replication.storage.openshift.io,resources=volumereplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=replication.storage.openshift.io,resources=volumereplications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *VolumeReplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	_ = r.Log.WithValues("controller_volumereplication", req.NamespacedName)

	// Fetch VolumeReplication instance
	instance := &replicationv1alpha1.VolumeReplication{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			r.Log.Info("no VolumeReplication resource found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Get VolumeReplicationClass
	vrcObj, err := r.getVolumeReplicaCLass(instance.Spec.VolumeReplicationClass)
	if err != nil {
		return ctrl.Result{}, err
	}

	if r.DriverConfig.DriverName != vrcObj.Spec.Provisioner {
		return ctrl.Result{}, nil
	}

	var volumeHandle string
	nameSpacedName := types.NamespacedName{Name: instance.Spec.DataSource.Name, Namespace: req.Namespace}
	switch instance.Spec.DataSource.Kind {
	case pvcDataSource:
		_, pv, err := r.getPVCDataSource(nameSpacedName)
		if err != nil {
			r.Log.Error(err, "failed to get dataSource for PVC", "dataSourceName", nameSpacedName.Name)
			return ctrl.Result{}, err
		}
		volumeHandle = pv.Spec.CSI.VolumeHandle
	default:
		return ctrl.Result{}, fmt.Errorf("unsupported datasource kind %q", instance.Spec.DataSource.Kind)
	}

	r.Log.Info("volume handle", volumeHandle)

	if instance.Spec.ImageState == replicationv1alpha1.Secondary {
		failedTask, err := markVolumeAsSecondary()
		if err != nil {
			r.Log.Error(err, "task failed", "taskName", failedTask)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReplicationReconciler) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {
	r.DriverConfig = cfg
	return ctrl.NewControllerManagedBy(mgr).
		For(&replicationv1alpha1.VolumeReplication{}).
		Complete(r)
}

// markVolumeAsSecondary defines and runs a set of tasks required to mark a volume as secondary
func markVolumeAsSecondary() (string, error) {
	var markVolumeAsSecondaryTasks = []*tasks.TaskSpec{
		{
			Name: "Demoting volume",
			Task: replication.NewDemoteVolumeTask(),
		},
		{
			Name: "Re-syncing volume",
			Task: replication.NewResyncVolumeTask(),
		},
	}

	return tasks.RunAll(markVolumeAsSecondaryTasks)
}
