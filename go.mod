module github.com/hashicorp/terraform-k8s

go 1.15

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/hashicorp/go-tfe v0.15.0
	github.com/hashicorp/terraform v0.14.3
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/stretchr/testify v1.6.1
	github.com/zclconf/go-cty v1.7.1
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v10.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/controller-tools v0.3.0 // indirect
)

// Pinned to kubernetes-1.16.2
//replace (
//	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191016112112-5190913f932d
//	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191016114015-74ad18325ed5
//	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191016115326-20453efc2458
//	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191016115129-c07a134afb42
//	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894
//	k8s.io/component-base => k8s.io/component-base v0.0.0-20191016111319-039242c015a9
//	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
//	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191016115521-756ffa5af0bd
//	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191016112429-9587704a8ad4
//	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191016114939-2b2b218dc1df
//	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191016114407-2e83b6f20229
//	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191016114748-65049c67a58b
//	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51
//	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191016114556-7841ed97f1b2
//	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191016115753-cf0698c3a16b
//	k8s.io/metrics => k8s.io/metrics v0.0.0-20191016113814-3b1a734dba6e
//	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191016112829-06bb3c9d77c9
//)

replace k8s.io/client-go => k8s.io/client-go v0.18.6
