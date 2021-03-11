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
	grpcClient "github.com/kube-storage/volume-replication-operator/pkg/client"
	"github.com/kube-storage/volume-replication-operator/pkg/config"
)

const (
	pvcDataSource            = "PersistentVolumeClaim"
	enableVolumeReplication  = "Enable volume replication"
	disableVolumeReplication = "Disable volume replication"
	promoteVolume            = "Promote volume"
	demoteVolume             = "Demote volume"
	resyncVolume             = "Resync volume"

	volumeReplicationFinalizer = "replication.storage.openshift.io"
)

// VolumeReplicationReconciler reconciles a VolumeReplication object
type VolumeReplicationReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	DriverConfig *config.DriverConfig
	GRPCClient   *grpcClient.Client
	Replication  grpcClient.VolumeReplication
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

	err = validatePrefixedParameters(vrcObj.Spec.Parameters)
	if err != nil {
		r.Log.Error(err, "failed to validate prefix parameters", "volumeReplicationClass", instance.Spec.VolumeReplicationClass)
		return ctrl.Result{}, err
	}
	// remove the prefix keys in volume replication class parameters
	parameters := filterPrefixedParameters(replicationParameterPrefix, vrcObj.Spec.Parameters)

	// get secret
	secretName := vrcObj.Spec.Parameters[prefixedReplicationSecretNameKey]
	secretNamespace := vrcObj.Spec.Parameters[prefixedReplicationSecretNamespaceKey]
	secret := make(map[string]string)
	if secretName != "" && secretNamespace != "" {
		secret, err = r.getSecret(secretName, secretNamespace)
		if err != nil {
			return reconcile.Result{}, err
		}
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

	// check if the object is being deleted
	if instance.GetDeletionTimestamp().IsZero() {
		if !contains(instance.GetFinalizers(), volumeReplicationFinalizer) {
			r.Log.Info("finalizer not found for volumeReplication object. Adding finalizer", "objectName", instance.Name, "finalizer", volumeReplicationFinalizer)
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, volumeReplicationFinalizer)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				r.Log.Error(err, "failed to update volumeReplication object with finalizer", "objectName", instance.Name, "finalizer", volumeReplicationFinalizer)
				return reconcile.Result{}, err
			}
		}
	} else {
		if contains(instance.GetFinalizers(), volumeReplicationFinalizer) {
			failedTask, err := r.disableVolumeReplication(volumeHandle, parameters, secret)
			if err != nil {
				r.Log.Error(err, "task failed", "taskName", failedTask)
				return ctrl.Result{}, err
			}

			r.Log.Info("removing finalizer from volumeReplication object", "objectName", instance.Name, "finalizer", volumeReplicationFinalizer)
			// once all finalizers have been removed, the object will be deleted
			instance.ObjectMeta.Finalizers = remove(instance.ObjectMeta.Finalizers, volumeReplicationFinalizer)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				r.Log.Error(err, "failed to remove finalizer from volumeReplication object", "objectName", instance.Name, "finalizer", volumeReplicationFinalizer)
				return reconcile.Result{}, err
			}
		}
		r.Log.Info("volumeReplication object is terminated, skipping reconciliation", "objectName", instance.Name)
		return ctrl.Result{}, nil
	}

	switch instance.Spec.ImageState {
	case replicationv1alpha1.Primary:
		failedTask, err := r.markVolumeAsPrimary(volumeHandle, parameters, secret)
		if err != nil {
			r.Log.Error(err, "task failed", "taskName", failedTask)
			return ctrl.Result{}, err
		}

	case replicationv1alpha1.Secondary:
		failedTask, err := r.markVolumeAsSecondary(volumeHandle, parameters, secret)
		if err != nil {
			r.Log.Error(err, "task failed", "taskName", failedTask)
			return ctrl.Result{}, err
		}

	case replicationv1alpha1.Resync:
		failedTask, err := r.resyncVolume(volumeHandle, parameters, secret)
		if err != nil {
			r.Log.Error(err, "task failed", "taskName", failedTask)
			return ctrl.Result{}, err
		}

	default:
		return ctrl.Result{}, fmt.Errorf("unsupported image state %q", instance.Spec.ImageState)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReplicationReconciler) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {
	r.DriverConfig = cfg
	c, err := grpcClient.New(cfg.DriverEndpoint, cfg.RPCTimeout)
	if err != nil {
		r.Log.Error(err, "failed to create GRPC Client", "Endpoint", cfg.DriverEndpoint, "GRPC Timeout", cfg.RPCTimeout)
		return err
	}
	err = c.Probe()
	if err != nil {
		r.Log.Error(err, "failed to connect to driver", "Endpoint", cfg.DriverEndpoint, "GRPC Timeout", cfg.RPCTimeout)
		return err
	}
	r.GRPCClient = c
	r.Replication = grpcClient.NewReplicationClient(r.GRPCClient.Client, cfg.RPCTimeout)
	return ctrl.NewControllerManagedBy(mgr).
		For(&replicationv1alpha1.VolumeReplication{}).
		Complete(r)
}

// markVolumeAsPrimary defines and runs a set of tasks required to mark a volume as primary
func (r *VolumeReplicationReconciler) markVolumeAsPrimary(volumeID string, parameters, secrets map[string]string) (string, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}
	var markVolumeAsPrimaryTasks = []*tasks.TaskSpec{
		{
			Name: enableVolumeReplication,
			Task: replication.NewEnableTask(c),
		},
		{
			Name: promoteVolume,
			Task: replication.NewPromoteVolumeTask(c),
		},
	}

	return tasks.RunAll(markVolumeAsPrimaryTasks)
}

// markVolumeAsSecondary defines and runs a set of tasks required to mark a volume as secondary
func (r *VolumeReplicationReconciler) markVolumeAsSecondary(volumeID string, parameters, secrets map[string]string) (string, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}
	var markVolumeAsSecondaryTasks = []*tasks.TaskSpec{
		{
			Name: demoteVolume,
			Task: replication.NewDemoteVolumeTask(c),
		},
		{
			Name: resyncVolume,
			Task: replication.NewResyncVolumeTask(c),
		},
	}

	return tasks.RunAll(markVolumeAsSecondaryTasks)
}

// resyncVolume defines and runs a set of tasks required to resync the volume
func (r *VolumeReplicationReconciler) resyncVolume(volumeID string, parameters, secrets map[string]string) (string, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}

	var resyncVolumeTasks = []*tasks.TaskSpec{
		{
			Name: resyncVolume,
			Task: replication.NewResyncVolumeTask(c),
		},
	}

	return tasks.RunAll(resyncVolumeTasks)
}

// disableVolumeReplication defines and runs a set of tasks required to disable volume replication
func (r *VolumeReplicationReconciler) disableVolumeReplication(volumeID string, parameters, secrets map[string]string) (string, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}

	var disableVolumeReplicationTasks = []*tasks.TaskSpec{
		{
			Name: disableVolumeReplication,
			Task: replication.NewDisableTask(c),
		},
	}

	return tasks.RunAll(disableVolumeReplicationTasks)
}
