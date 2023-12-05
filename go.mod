module github.com/snyk/terraform-k8s

go 1.16

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.6 // indirect
	github.com/hashicorp/go-tfe v0.21.0
	github.com/hashicorp/terraform v0.15.2
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/zclconf/go-cty v1.9.1
	golang.org/x/net v0.0.0-20211020060615-d418f374d309 // indirect
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1 // indirect
	golang.org/x/sys v0.0.0-20211025201205-69cdffdb9359 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
	k8s.io/api v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v10.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.9.7
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.5.5
	github.com/hashicorp/consul/api => github.com/hashicorp/consul/api v1.11.0
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common => github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.277
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag => github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag v1.0.277
	k8s.io/client-go => k8s.io/client-go v0.21.4
)
