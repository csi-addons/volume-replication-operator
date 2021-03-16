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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	replicationlib "github.com/kube-storage/spec/lib/go/replication"
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
	forcePromoteVolume       = "Promote volume (forced)"
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
			err := r.disableVolumeReplication(volumeHandle, parameters, secret)
			if err != nil {
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
		err := r.markVolumeAsPrimary(instance, volumeHandle, parameters, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

	case replicationv1alpha1.Secondary:
		requeueForResync, err := r.markVolumeAsSecondary(instance, volumeHandle, parameters, secret)
		if err != nil {
			return ctrl.Result{}, err
		}
		if requeueForResync {
			return ctrl.Result{Requeue: true}, nil
		}

	case replicationv1alpha1.Resync:
		requeueForResync, err := r.resyncVolume(instance, volumeHandle, parameters, secret)
		if err != nil {
			return ctrl.Result{}, err
		}
		if requeueForResync {
			return ctrl.Result{Requeue: true}, nil
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
func (r *VolumeReplicationReconciler) markVolumeAsPrimary(volumeReplicationObject *replicationv1alpha1.VolumeReplication, volumeID string, parameters, secrets map[string]string) error {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}
	var err error

	if volumeReplicationObject.Status.State != replicationv1alpha1.Replicating {
		enableVolumeReplicationTask := []*tasks.TaskSpec{
			{
				Name: enableVolumeReplication,
				Task: replication.NewEnableTask(c),
			},
		}
		resp := tasks.RunAll(enableVolumeReplicationTask)
		for _, re := range resp {
			if re.Error != nil {
				r.Log.Error(re.Error, "task failed", "taskName", re.Name)
				return re.Error
			}
		}
		volumeReplicationObject.Status.State = replicationv1alpha1.Replicating
		err = r.Client.Status().Update(context.TODO(), volumeReplicationObject, nil)
		if err != nil {
			return err
		}
	}

	promoteVolumeTasks := []*tasks.TaskSpec{
		{
			Name: promoteVolume,
			Task: replication.NewPromoteVolumeTask(c, false),
			KnownErrors: []codes.Code{
				codes.FailedPrecondition,
			},
		},
	}
	resp := tasks.RunAll(promoteVolumeTasks)

	isKnownError := r.hasKnownGRPCError(promoteVolumeTasks, resp)
	if isKnownError {
		forcePromoteVolumeTasks := []*tasks.TaskSpec{
			{
				Name: forcePromoteVolume,
				Task: replication.NewPromoteVolumeTask(c, true),
			},
		}
		resp := tasks.RunAll(forcePromoteVolumeTasks)
		for _, re := range resp {
			if re.Error != nil {
				r.Log.Error(re.Error, "task failed", "taskName", re.Name)
				return re.Error
			}
		}
	}

	return nil
}

// markVolumeAsSecondary defines and runs a set of tasks required to mark a volume as secondary
func (r *VolumeReplicationReconciler) markVolumeAsSecondary(volumeReplicationObject *replicationv1alpha1.VolumeReplication,
	volumeID string, parameters, secrets map[string]string) (bool, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}
	var err error

	// if we are not re-syncing, we need to demote volume first
	if volumeReplicationObject.Status.State != replicationv1alpha1.Resyncing {
		demoteVolumeTask := []*tasks.TaskSpec{
			{
				Name: demoteVolume,
				Task: replication.NewDemoteVolumeTask(c),
			},
		}
		resp := tasks.RunAll(demoteVolumeTask)
		for _, re := range resp {
			if re.Error != nil {
				r.Log.Error(re.Error, "task failed", "taskName", re.Name)
				return false, re.Error
			}
		}
		// set status.state to re-syncing before moving forward
		volumeReplicationObject.Status.State = replicationv1alpha1.Resyncing
		err = r.Client.Status().Update(context.TODO(), volumeReplicationObject, nil)
		if err != nil {
			return false, err
		}
	}

	// if we are re-syncing, we need to check response for Ready and requeue if required
	if volumeReplicationObject.Status.State == replicationv1alpha1.Resyncing {
		return r.resyncVolume(volumeReplicationObject, volumeID, parameters, secrets)
	}

	return false, nil
}

// resyncVolume defines and runs a set of tasks required to resync the volume
func (r *VolumeReplicationReconciler) resyncVolume(volumeReplicationObject *replicationv1alpha1.VolumeReplication,
	volumeID string, parameters, secrets map[string]string) (bool, error) {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}
	var err error

	if volumeReplicationObject.Status.State != replicationv1alpha1.Resyncing {
		volumeReplicationObject.Status.State = replicationv1alpha1.Resyncing
		err = r.Client.Status().Update(context.TODO(), volumeReplicationObject, nil)
		if err != nil {
			return false, err
		}
	}

	var resyncVolumeTasks = []*tasks.TaskSpec{
		{
			Name: resyncVolume,
			Task: replication.NewResyncVolumeTask(c),
		},
	}

	resp := tasks.RunAll(resyncVolumeTasks)
	for _, re := range resp {
		if re.Error != nil {
			r.Log.Error(re.Error, "task failed", "taskName", re.Name)
			return false, re.Error
		}
		resyncResponse, ok := re.Response.(replicationlib.ResyncVolumeResponse)
		if !ok {
			err = fmt.Errorf("received response of unexpected type")
			r.Log.Error(err, "unable to parse response")
			return false, err
		}
		if !resyncResponse.GetReady() {
			return true, nil
		}
	}
	volumeReplicationObject.Status.State = replicationv1alpha1.Replicating
	err = r.Client.Status().Update(context.TODO(), volumeReplicationObject, nil)
	if err != nil {
		return false, err
	}
	return false, nil
}

// disableVolumeReplication defines and runs a set of tasks required to disable volume replication
func (r *VolumeReplicationReconciler) disableVolumeReplication(volumeID string, parameters, secrets map[string]string) error {
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

	resp := tasks.RunAll(disableVolumeReplicationTasks)
	// Check error for all tasks and return error
	for _, re := range resp {
		if re.Error != nil {
			r.Log.Error(re.Error, "task failed", "taskName", re.Name)
			return re.Error
		}
	}
	return nil
}

func (r *VolumeReplicationReconciler) hasKnownGRPCError(tasks []*tasks.TaskSpec, responses []*tasks.TaskResponse) bool {
	for _, re := range responses {
		if re.Error != nil {
			s, ok := status.FromError(re.Error)
			if !ok {
				// This is not gRPC error. The operation must have failed before gRPC
				// method was called, otherwise we would get gRPC error.
				r.Log.Error(re.Error, "task failed", "taskName", re.Name)
				return false
			}
			for _, task := range tasks {
				for _, e := range task.KnownErrors {
					if s.Code() == e {
						return true
					}
				}
			}
			r.Log.Error(re.Error, "task failed", "taskName", re.Name)
			return false
		}
	}
	return false
}
