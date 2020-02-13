# Syncing Kubernetes and Terraform Cloud Workspaces

By creating the definition of a Workspace Custom Resource in Kubernetes,
workspace definitions in Kubernetes can be automatically synced to those in
Terraform Cloud by the Kubernetes [Operator
pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/). This
functionality is provided by the [terraform-k8s
project](https://github.com/hashicorp/terraform-k8s) and can be automatically
installed and configured using the [Terraform Helm
chart](https://github.com/hashicorp/terraform-helm).

**Why create a Terraform Cloud Workspace Custom Resource Definition for
Kubernetes?** Applications deployed in Kubernetes will have the ability to
define infrastructure configuration using a Workspace in Kubernetes. The
functionality depends on Terraform Cloud to ensure consistent approaches to
state locking, state storage, and execution.

**Why sync Kubernetes workspaces to Terraform Cloud?** This functionality will
re-execute updates to infrastructure and Terraform Cloud non-sensitive
variables. It provides a first-class Kubernetes interface to updating Terraform
Cloud.

**How does it work?** The workspace sync is done using a Terraform Operator in
the [terraform-k8s project](https://github.com/hashicorp/terraform-k8s). The
Terraform Operator must run within a Kubernetes cluster and be scoped to a
namespace. The deployment of the operator and resource definition is automated
by a [Helm chart](https://github.com/hashicorp/terraform-helm).

**How does the operator handle sensitive variables for Terraform Cloud?** There
are two categories of sensitive variables related to Terraform Cloud:
1. Terraform Cloud API Token: used to log in and execute runs for Terraform
   Cloud
2. Workspace Sensitive Variables: secrets that execution requires to log into
   providers (e.g., credentials).
   
See the [Authentication](#authentication) and
[Workspace Sensitive Variables](#workspace-sensitive-variables) sections for
how these are handled.

## Installation and Configuration

For a full set of Kubernetes manifests related to the operator, see the
`operator/deploy/` directory. Otherwise, a fully automated deployment can be
found at the [Helm chart](https://github.com/hashicorp/terraform-helm).

### Namespace

Create the namespace to which you would like to deploy the Operator, Secrets,
and Workspace resources.

```shell
> kubectl create ns $NAMESPACE
```

### Custom Resource Definition

Before running the operator, you must deploy the Custom Resource Definition for
the Workspace. Custom Resource Definitions extend the Kubernetes API.

```shell
> kubectl apply -f operator/deploy/crds/app.terraform.io_workspaces_crd.yaml
```

The Custom Resource Definition defines that the Workspace Custom Resource
schema.

### Role-Based Access Control

In order to scope the operator to a namespace, you must assign a role and
service account to the namespace.

```shell
> kubectl -n $NAMESPACE apply -f operator/deploy/role_binding.yaml
> kubectl -n $NAMESPACE apply -f operator/deploy/role.yaml
> kubectl -n $NAMESPACE apply -f operator/deploy/service_account.yaml
```

Generally, the role must have access to Pods, Secrets, Services, and ConfigMaps.

### Authentication

The operator must authenticate to Terraform Cloud. Note that `terraform-k8s`
must run within the cluster, which means already handles Kubernetes
authentication. The Terraform Cloud API token can be generated under
`https://app.terraform.io/app/${ORGANIZATION}/settings/teams`. Once generating
this token, insert it into a file formatted for Terraform credentials.

```hcl
credentials app.terraform.io {
  token = "${TERRAFORM_CLOUD_API_TOKEN}"
}
```

Note that the Terraform Cloud API token is a broad-spectrum token. It allows
creation of workspaces and execution of runs. It will not offer access control
on the workspace level or roles within the team. In order to support a
first-class Kubernetes experience, security and access control to this token
must be enforced by [Kubernetes Role-Based Access Control
(RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) policies.

As long as there is an environment variable called `TF_CLI_CONFIG_FILE` and the
entire credentials file with the Terraform Cloud API Token is mounted to the
filepath specified by `TF_CLI_CONFIG_FILE`, the operator will be able to
authenticate against Terraform Cloud. This supports secrets management
approaches in Kubernetes that use a volume mount for secrets.

For example, we create a Kubernetes secret and mount it to the operator at the
filepath specified by `TF_CLI_CONFIG_FILE`.

```yaml
---
# not secure secrets management
apiVersion: apps/v1
kind: Secret
metadata:
  name: terraformrc
type: Opaque
data:
  credentials: |-
    credentials app.terraform.io {
      token = "${TERRAFORM_CLOUD_API_TOKEN}"
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: terraform-k8s
spec:
  # some sections omitted for clarity
  template:
    metadata:
      labels:
        name: terraform-k8s
    spec:
      serviceAccountName: terraform-k8s
      containers:
        - name: terraform-k8s
          env:
            - name: TF_CLI_CONFIG_FILE
              value: "/etc/terraform/.terraformrc"
          volumeMounts:
          - name: terraformrc
            mountPath: "/etc/terraform"
            readOnly: true
      volumes:
        - name: terraformrc
          secret:
            secretName: terraformrc
            items:
            - key: credentials
              path: ".terraformrc"
```

### Workspace Sensitive Variables

Sensitive variables in Terraform Cloud workspaces often take the form of
credentials for cloud providers or API endpoints. They enable Terraform Cloud to
authenticate against a provider and apply changes to infrastructure.

Instead of defining sensitive variables in a Workspace CustomResource, you mount
them to the operator's deployment for use. This ensures that only the operator
has access to read and create sensitive variables as part of the Terraform Cloud
workspace.

Similar to the Terraform Cloud API Token, workspace sensitive variables must be
mounted on the operator deployment as a directory. The secret name should map to
the file name of the secret and the value should be the file's contents. This
supports secrets management approaches in Kubernetes that use a volume mount for
secrets.

```yaml
---
# not secure secrets management
apiVersion: apps/v1
kind: Secret
metadata:
  name: workspacesecrets
type: Opaque
data:
  AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY}
  GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: terraform-k8s
spec:
  # some sections omitted for clarity
  template:
    metadata:
      labels:
        name: terraform-k8s
    spec:
      serviceAccountName: terraform-k8s
      containers:
        - name: terraform-k8s
          volumeMounts:
          - name: workspacesecrets
            mountPath: "/tmp/secrets"
            readOnly: true
      volumes:
        - name: workspacesecrets
          secret:
            secretName: workspacesecrets
```

 In order to support a first-class Kubernetes experience, security and access
 control to these secrets must be enforced by [Kubernetes Role-Based Access
 Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
 policies.

### Namespace Scope

To ensure the operator does not have access to secrets or resource beyond the
namespace, the deployment must be namespace-scoped.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: terraform-k8s
spec:
  # some sections omitted for clarity
  template:
    metadata:
      labels:
        name: terraform-k8s
    spec:
      serviceAccountName: terraform-k8s
      containers:
        - name: terraform-k8s
          command:
          - terraform-k8s
          - sync-workspace
          - "--k8s-watch-namespace=$(POD_NAMESPACE)"
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

When deploying, ensure that the namespace is passed into the
`--k8s-watch-namespace` option. Otherwise, the operator will attempt to access
across all namespaces (cluster scope).

### Workspace (CustomResource)

The Workspace CustomResource defines a Terraform Cloud workspace, including
variables, Terraform module, and outputs.

For a complete example of a Workspace CustomResource, see
`operator/deploy/crds/app.terraform.io_v1alpha1_workspace_cr.yaml`.

The Workspace Spec includes the following parameters:

1. `organization`: The Terraform Cloud organization you would like to use.

1. `secretsMountPath`: The file path defined on the operator deployment that
   contains the workspace's secrets.

Additional parameters are outlined below. Changes to `organization`, `module`,
`outputs`, and non-sensitive `variables` will trigger a new execution in
Terraform Cloud.

#### Modules

> The Workspace will only execute Terraform configuration in a module. It will
> not execute `*.tf` files.

Information passed to the Workspace CustomResource will be rendered to a
template Terraform configuration that uses the `module` block. Specify a module
with remote `source`. Publicly available VCS repositories, the Terraform
Registry, and private module registry are supported. In addition to `source`,
specify a module `version`.

```yaml
module:
  source: "hashicorp/hello/random"
  version: "3.1.0"
```

#### Outputs

In order to map the Terraform output to the module output, specify the `outputs`
section of the Workspace CustomResource. The `key` denotes the output that you
expect from `terraform output` and the value will map to the `moduleOutputName`
specified.

```yaml
outputs:
  - key: my_pet
    moduleOutputName: pet
```

#### Variables

In general, updates to **non-sensitive** variables will trigger a new execution
in Terraform Cloud. However, updates to sensitive variables *will not* trigger a
new execution. This is because sensitive variables are write-only for security
purposes and the operator is unable to reconcile the upstream value of the
secret with the value stored locally.

You can define Terraform variables in two ways:

1. Inline
   ```yaml
   variables:
     - key: hello
       value: rosemary
       sensitive: false
       environmentVariable: false
   ```

2. With a Kubernetes ConfigMap reference
   ```yaml
   variables:
     - key: second_hello
       valueFrom:
         configMapKeyRef:
           name: say-hello
           key: to
       sensitive: false
       environmentVariable: false
   ```

You must include whether or not the variable is `sensitive` (a secret) or should
be defined as an `environmentVariable`.

For variables that are secret, they should already be initialized as per
[Workspace Sensitive Variables](#workspace-sensitive-variables). You can define
them by setting `sensitive: true`. Do not define the value or use a ConfigMap
reference, as the read from file will override the value you set.

```yaml
variables:
  - key: AWS_SECRET_ACCESS_KEY
    sensitive: true
    environmentVariable: true
```

### Workspace Destruction

In order for workspace destruction to work automatically, you must set the
`CONFIRM_DESTROY` environment variable in the Terraform Cloud workspace. When
you delete the Workspace CustomResource, the operator will attempt to destroy
the workspace. As a secondary check, you must deploy the operator with this
environment variable defined in the `variables` section if you would like to
destroy the workspace in Terraform Cloud.

```yaml
variables:
  - key: CONFIRM_DESTROY
    value: "1"
    sensitive: false
    environmentVariable: true
```

When deleting the Workspace CustomResource, the command line will wait for
a few moments.

```shell
> kubectl delete workspace.app.terraform.io/my-workspace
```

This is because the operator is running a
[finalizer](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers).
The finalizer will execute before the workspace officially deletes in order to:

1. Stop all runs in the workspace, including pending ones
1. `terraform destroy -auto-approve` on resources in the workspace
1. Delete the workspace.

Once the finalizer completes, Kubernetes deletes the Workspace CustomResource.