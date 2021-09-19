# Volume Replication Operator

## Overview

Volume Replication Operator is a kubernetes operator that provides common and reusable APIs for storage disaster recovery.
It is based on [csi-addons/spec](https://github.com/csi-addons/spec) specification and can be used by any storage
provider.

## Design

Volume Replication Operator follows controller pattern and provides extended APIs for storage disaster recovery.
The extended APIs are provided via Custom Resource Definition (CRD).

### [VolumeReplicationClass](https://github.com/csi-addons/volume-replication-operator/blob/main/config/crd/bases/replication.storage.openshift.io_volumereplicationclasses.yaml)

`VolumeReplicationClass` is a cluster scoped resource that contains driver related configuration parameters.

`provisioner` is name of the storage provisioner

`parameters` contains key-value pairs that are passed down to the driver. Users can add their own key-value pairs.
Keys with `replication.storage.openshift.io/` prefix are reserved by operator and not passed down to the driver.

#### Reserved parameter keys

+ `replication.storage.openshift.io/replication-secret-name`
+ `replication.storage.openshift.io/replication-secret-namespace`

```yaml
apiVersion: replication.storage.openshift.io/v1alpha1
kind: VolumeReplicationClass
metadata:
  name: volumereplicationclass-sample
spec:
  provisioner: example.provisioner.io
  parameters:
    replication.storage.openshift.io/replication-secret-name: secret-name
    replication.storage.openshift.io/replication-secret-namespace: secret-namespace
```

### [VolumeReplication](https://github.com/csi-addons/volume-replication-operator/blob/main/config/crd/bases/replication.storage.openshift.io_volumereplications.yaml)

VolumeReplication is a namespaced resource that contains references to storage object to be replicated and
VolumeReplicationClass corresponding to the driver providing replication.

`volumeReplicationClass` is the class providing replication

`replicationState` is the state of the volume being referenced. Possible values are `primary`. `secondary` and `resync`.
  + `primary` denotes that the volume is primary
  + `secondary` denotes that the volume is secondary
  + `resync` denotes that the volume needs to be resynced

`dataSource` contains typed reference to the source being replicated.
  + `apiGroup` is the group for the resource being referenced. If apiGroup is not specified, the specified Kind must
  be in the core API group. For any other third-party types, apiGroup is required.
  + `kind` is the kind of resource being replicated. For eg. `PersistentVolumeClaim`
  + `name` is the name of the resource

`replicationHandle` (optional) is an existing (but new) replication id

```yaml
apiVersion: replication.storage.openshift.io/v1alpha1
kind: VolumeReplication
metadata:
  name: volumereplication-sample
  namespace: default
spec:
  volumeReplicationClass: volumereplicationclass-sample
  replicationState: primary
  replicationHandle: replicationHandle # optional
  dataSource:
    kind: PersistentVolumeClaim
    name: myPersistentVolumeClaim # should be in same namespace as VolumeReplication
```

## Usage

### Planned Storage Migration

#### Failover

In case of planned migration, update `replicationState` to `secondary` in `VolumeReplication` CR at Primary Site. When the operator sees this change, it will pass the information down to the driver via GRPC request to mark the `dataSource` as secondary.

On the Secondary Site, create `VolumeReplication` CR pointing to the same `dataSource` with `replicationState` as `primary`. When the operator sees this change, it will pass the information down to the driver via GRPC request to mark the dataSource as primary.

#### Failback

Once the planned work is done and you want to failback, update `replicationState` from `primary` to `secondary` in current Primary Site and `secondary` to `primary` in current Secondary Site after it is recovered. These changes are detected by operator and information is passed down to the driver via GRPC request.

### Storage Disaster Recovery

#### Failover

In case of disaster recovery, create `VolumeReplication` CR at Secondary Site. Since the connection to the Primary Site is lost, the operator automatically sends a GRPC request down to the driver to forcefully mark the dataSource as primary.

#### Failback

Once the failed cluster is recovered and you want to failback, update `replicationState` from `primary` to `secondary` in current Primary Site and `secondary` to `primary` in the recovered site. These changes are detected by operator and information is passed down to the driver via GRPC request to make the necessary changes.
