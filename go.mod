module github.com/csi-addons/volume-replication-operator

go 1.15

require (
	github.com/csi-addons/spec v0.1.1
	github.com/go-logr/logr v0.4.0
	github.com/kubernetes-csi/csi-lib-utils v0.9.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.18.1
	google.golang.org/grpc v1.32.0
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.9.6
)
