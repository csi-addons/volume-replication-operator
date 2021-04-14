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
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	replicationlib "github.com/csi-addons/spec/lib/go/replication"
	replicationv1alpha1 "github.com/csi-addons/volume-replication-operator/api/v1alpha1"
	"github.com/csi-addons/volume-replication-operator/controllers/tasks"
	"github.com/csi-addons/volume-replication-operator/controllers/tasks/replication"
	grpcClient "github.com/csi-addons/volume-replication-operator/pkg/client"
	"github.com/csi-addons/volume-replication-operator/pkg/config"
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

	logger := r.Log.WithValues("Request.Name", req.Name, "Request.Namespace", req.Namespace)

	// Fetch VolumeReplication instance
	instance := &replicationv1alpha1.VolumeReplication{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("volumeReplication resource not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Get VolumeReplicationClass
	vrcObj, err := r.getVolumeReplicaCLass(logger, instance.Spec.VolumeReplicationClass)
	if err != nil {
		setFailureCondition(instance)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
		return ctrl.Result{}, err
	}

	if r.DriverConfig.DriverName != vrcObj.Spec.Provisioner {
		return ctrl.Result{}, nil
	}

	err = validatePrefixedParameters(vrcObj.Spec.Parameters)
	if err != nil {
		setFailureCondition(instance)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
		logger.Error(err, "failed to validate parameters of volumeReplicationClass", "VRCName", instance.Spec.VolumeReplicationClass)
		return ctrl.Result{}, err
	}
	// remove the prefix keys in volume replication class parameters
	parameters := filterPrefixedParameters(replicationParameterPrefix, vrcObj.Spec.Parameters)

	// get secret
	secretName := vrcObj.Spec.Parameters[prefixedReplicationSecretNameKey]
	secretNamespace := vrcObj.Spec.Parameters[prefixedReplicationSecretNamespaceKey]
	secret := make(map[string]string)
	if secretName != "" && secretNamespace != "" {
		secret, err = r.getSecret(logger, secretName, secretNamespace)
		if err != nil {
			setFailureCondition(instance)
			_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
			return reconcile.Result{}, err
		}
	}

	var volumeHandle string
	nameSpacedName := types.NamespacedName{Name: instance.Spec.DataSource.Name, Namespace: req.Namespace}
	switch instance.Spec.DataSource.Kind {
	case pvcDataSource:
		_, pv, err := r.getPVCDataSource(logger, nameSpacedName)
		if err != nil {
			setFailureCondition(instance)
			_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
			logger.Error(err, "failed to get PVC", "PVCName", instance.Spec.DataSource.Name)
			return ctrl.Result{}, err
		}
		volumeHandle = pv.Spec.CSI.VolumeHandle
	default:
		err = fmt.Errorf("unsupported datasource kind")
		logger.Error(err, "given kind not supported", "Kind", instance.Spec.DataSource.Kind)
		setFailureCondition(instance)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
		return ctrl.Result{}, nil
	}

	logger.Info("volume handle", "VolumeHandleName", volumeHandle)

	// check if the object is being deleted
	if instance.GetDeletionTimestamp().IsZero() {
		if !contains(instance.GetFinalizers(), volumeReplicationFinalizer) {
			logger.Info("finalizer not found for volumeReplication object. Adding finalizer", "Finalizer", volumeReplicationFinalizer)
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, volumeReplicationFinalizer)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				logger.Error(err, "failed to update volumeReplication object with finalizer", "Finalizer", volumeReplicationFinalizer)
				return reconcile.Result{}, err
			}
		}
	} else {
		if contains(instance.GetFinalizers(), volumeReplicationFinalizer) {
			err := r.disableVolumeReplication(logger, volumeHandle, parameters, secret)
			if err != nil {
				logger.Error(err, "failed to disable replication")
				return ctrl.Result{}, err
			}

			logger.Info("removing finalizer from volumeReplication object", "Finalizer", volumeReplicationFinalizer)
			// once all finalizers have been removed, the object will be deleted
			instance.ObjectMeta.Finalizers = remove(instance.ObjectMeta.Finalizers, volumeReplicationFinalizer)
			if err := r.Client.Update(context.TODO(), instance); err != nil {
				logger.Error(err, "failed to remove finalizer from volumeReplication object", "Finalizer", volumeReplicationFinalizer)
				return reconcile.Result{}, err
			}
		}
		logger.Info("volumeReplication object is terminated, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	instance.Status.LastStartTime = getCurrentTime()
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		logger.Error(err, "failed to update status")
		return reconcile.Result{}, err
	}

	// enable replication on every reconcile
	if err = r.enableReplication(logger, volumeHandle, parameters, secret); err != nil {
		logger.Error(err, "failed to enable replication")
		setFailureCondition(instance)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), err.Error())
		return reconcile.Result{}, err
	}

	var replicationErr error
	var requeueForResync bool

	switch instance.Spec.ReplicationState {
	case replicationv1alpha1.Primary:
		replicationErr = r.markVolumeAsPrimary(instance, logger, volumeHandle, parameters, secret)

	case replicationv1alpha1.Secondary:
		replicationErr = r.markVolumeAsSecondary(instance, logger, volumeHandle, parameters, secret)
		// resync volume if successfully marked Secondary
		if replicationErr == nil {
			err := r.updateReplicationStatus(instance, logger, getReplicationState(instance), "volume is marked secondary")
			if err != nil {
				return ctrl.Result{}, err
			}
			requeueForResync, replicationErr = r.resyncVolume(instance, logger, volumeHandle, parameters, secret)
		}

	case replicationv1alpha1.Resync:
		requeueForResync, replicationErr = r.resyncVolume(instance, logger, volumeHandle, parameters, secret)

	default:
		replicationErr = fmt.Errorf("unsupported volume state")
		logger.Error(replicationErr, "given volume state is not supported", "ReplicationState", instance.Spec.ReplicationState)
		setFailureCondition(instance)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), replicationErr.Error())
		return ctrl.Result{}, nil
	}

	if replicationErr != nil {
		logger.Error(replicationErr, "failed to Replicate", "ReplicationState", instance.Spec.ReplicationState)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), replicationErr.Error())
		return ctrl.Result{}, replicationErr
	}

	if requeueForResync {
		logger.Info("volume is not ready to use, requeuing for resync")
		setDegradedCondition(&instance.Status.Conditions, instance.Generation)
		_ = r.updateReplicationStatus(instance, logger, getCurrentReplicationState(instance), "volume is degraded")
		return ctrl.Result{Requeue: true}, nil
	}

	var msg string
	if instance.Spec.ReplicationState == replicationv1alpha1.Resync {
		msg = "volume is marked for resyncing"
	} else {
		msg = fmt.Sprintf("volume is marked %s", string(instance.Spec.ReplicationState))
	}

	instance.Status.LastCompletionTime = getCurrentTime()
	err = r.updateReplicationStatus(instance, logger, getReplicationState(instance), msg)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info(msg)

	return ctrl.Result{}, nil
}

func (r *VolumeReplicationReconciler) updateReplicationStatus(instance *replicationv1alpha1.VolumeReplication, logger logr.Logger, state replicationv1alpha1.State, message string) error {
	instance.Status.State = state
	instance.Status.Message = message
	instance.Status.ObservedGeneration = instance.Generation
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
		logger.Error(err, "failed to update status")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VolumeReplicationReconciler) SetupWithManager(mgr ctrl.Manager, cfg *config.DriverConfig) error {

	pred := predicate.GenerationChangedPredicate{}

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
		WithEventFilter(pred).Complete(r)
}

// markVolumeAsPrimary defines and runs a set of tasks required to mark a volume as primary
func (r *VolumeReplicationReconciler) markVolumeAsPrimary(volumeReplicationObject *replicationv1alpha1.VolumeReplication, logger logr.Logger, volumeID string, parameters, secrets map[string]string) error {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
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

	isKnownError := r.hasKnownGRPCError(logger, promoteVolumeTasks, resp)

	if !isKnownError {
		for _, re := range resp {
			if re.Error != nil {
				logger.Error(re.Error, "task failed", "taskName", re.Name)
				setFailedPromotionCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
				return re.Error
			}
		}

	} else {
		forcePromoteVolumeTasks := []*tasks.TaskSpec{
			{
				Name: forcePromoteVolume,
				Task: replication.NewPromoteVolumeTask(c, true),
			},
		}
		resp := tasks.RunAll(forcePromoteVolumeTasks)
		for _, re := range resp {
			if re.Error != nil {
				logger.Error(re.Error, "task failed", "taskName", re.Name)
				setFailedPromotionCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
				return re.Error
			}
		}
	}

	setPromotedCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
	return nil
}

// markVolumeAsSecondary defines and runs a set of tasks required to mark a volume as secondary
func (r *VolumeReplicationReconciler) markVolumeAsSecondary(volumeReplicationObject *replicationv1alpha1.VolumeReplication,
	logger logr.Logger, volumeID string, parameters, secrets map[string]string) error {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}

	demoteVolumeTask := []*tasks.TaskSpec{
		{
			Name: demoteVolume,
			Task: replication.NewDemoteVolumeTask(c),
		},
	}
	resp := tasks.RunAll(demoteVolumeTask)
	for _, re := range resp {
		if re.Error != nil {
			logger.Error(re.Error, "task failed", "taskName", re.Name)
			setFailedDemotionCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
			return re.Error
		}
	}

	setDemotedCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
	return nil
}

// resyncVolume defines and runs a set of tasks required to resync the volume
func (r *VolumeReplicationReconciler) resyncVolume(volumeReplicationObject *replicationv1alpha1.VolumeReplication,
	logger logr.Logger, volumeID string, parameters, secrets map[string]string) (bool, error) {
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

	resp := tasks.RunAll(resyncVolumeTasks)
	for _, re := range resp {
		if re.Error != nil {
			logger.Error(re.Error, "task failed", "taskName", re.Name)
			setFailedResyncCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
			return false, re.Error
		}
		resyncResponse, ok := re.Response.(*replicationlib.ResyncVolumeResponse)
		if !ok {
			err := fmt.Errorf("received response of unexpected type")
			logger.Error(err, "unable to parse response")
			setFailedResyncCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
			return false, err
		}
		if !resyncResponse.GetReady() {
			return true, nil
		}
	}

	setResyncCondition(&volumeReplicationObject.Status.Conditions, volumeReplicationObject.Generation)
	return false, nil
}

// disableVolumeReplication defines and runs a set of tasks required to disable volume replication
func (r *VolumeReplicationReconciler) disableVolumeReplication(logger logr.Logger, volumeID string, parameters, secrets map[string]string) error {
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
			KnownErrors: []codes.Code{
				codes.NotFound,
			},
		},
	}

	resp := tasks.RunAll(disableVolumeReplicationTasks)

	if isKnownError := r.hasKnownGRPCError(logger, disableVolumeReplicationTasks, resp); isKnownError {
		logger.Info("volume not found", "volumeID", volumeID)
		return nil
	}

	// Check error for all tasks and return error
	for _, re := range resp {
		if re.Error != nil {
			logger.Error(re.Error, "task failed", "taskName", re.Name)
			return re.Error
		}
	}
	return nil
}

// enableReplication enable volume replication on the first reconcile
func (r *VolumeReplicationReconciler) enableReplication(logger logr.Logger, volumeID string, parameters, secrets map[string]string) error {
	c := replication.CommonRequestParameters{
		VolumeID:    volumeID,
		Parameters:  parameters,
		Secrets:     secrets,
		Replication: r.Replication,
	}

	enableVolumeReplicationTask := []*tasks.TaskSpec{
		{
			Name: enableVolumeReplication,
			Task: replication.NewEnableTask(c),
		},
	}
	resp := tasks.RunAll(enableVolumeReplicationTask)
	for _, re := range resp {
		if re.Error != nil {
			logger.Error(re.Error, "task failed", "taskName", re.Name)
			return re.Error
		}
	}
	return nil
}

func (r *VolumeReplicationReconciler) hasKnownGRPCError(logger logr.Logger, tasks []*tasks.TaskSpec, responses []*tasks.TaskResponse) bool {
	for _, re := range responses {
		if re.Error != nil {
			s, ok := status.FromError(re.Error)
			if !ok {
				// This is not gRPC error. The operation must have failed before gRPC
				// method was called, otherwise we would get gRPC error.
				logger.Error(re.Error, "task failed", "taskName", re.Name)
				return false
			}
			for _, task := range tasks {
				for _, e := range task.KnownErrors {
					if s.Code() == e {
						return true
					}
				}
			}
			logger.Error(re.Error, "task failed", "taskName", re.Name)
			return false
		}
	}
	return false
}

func getReplicationState(instance *replicationv1alpha1.VolumeReplication) replicationv1alpha1.State {
	switch instance.Spec.ReplicationState {
	case replicationv1alpha1.Primary:
		return replicationv1alpha1.PrimaryState
	case replicationv1alpha1.Secondary:
		return replicationv1alpha1.SecondaryState
	case replicationv1alpha1.Resync:
		return replicationv1alpha1.SecondaryState
	}

	return replicationv1alpha1.UnknownState
}

func getCurrentReplicationState(instance *replicationv1alpha1.VolumeReplication) replicationv1alpha1.State {
	if instance.Status.State == "" {
		return replicationv1alpha1.UnknownState
	}
	return instance.Status.State
}

func setFailureCondition(instance *replicationv1alpha1.VolumeReplication) {
	switch instance.Spec.ReplicationState {
	case replicationv1alpha1.Primary:
		setFailedPromotionCondition(&instance.Status.Conditions, instance.Generation)
	case replicationv1alpha1.Secondary:
		setFailedDemotionCondition(&instance.Status.Conditions, instance.Generation)
	case replicationv1alpha1.Resync:
		setFailedResyncCondition(&instance.Status.Conditions, instance.Generation)
	}
}

func getCurrentTime() *metav1.Time {
	metav1NowTime := metav1.NewTime(time.Now())
	return &metav1NowTime
}
