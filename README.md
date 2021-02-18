# Terraform Cloud Operator for Kubernetes

The Terraform Cloud Operator for Kubernetes provides first-class integration between [Kubernetes](https://kubernetes.io/) and [Terraform Cloud](https://www.hashicorp.com/products/terraform/editions/cloud/) by extending the Kubernetes control plane to enable lifecycle management of cloud and on-prem infrastructure through [Kubernetes manifests](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/). Manifests can be deployed and managed using kubectl, [Terraform](https://registry.terraform.io/providers/hashicorp/kubernetes-alpha/latest), Gitops tools, or any other tool that allows you to manage Kubernetes [custom resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). 

This operator provides a unified way to manage a Kubernetes application and its infrastructure dependencies through a single Kubernetes [CustomResourceDefinition](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/) (CRD). After the infrastructure dependencies are created, pertinent information such as endpoints and credentials are returned from Terraform Cloud to Kubernetes.

## Use Case
  * **Manage the lifecycle of cloud and on-prem infrastructure through a single Kubernetes custom resource**  
    * Install the operator from the corresponding [Helm Chart](https://github.com/hashicorp/terraform-helm) to enable the management of infrastructure services from any Kubernetes cluster.
    * Provision and manage infrastructure from any provider, such as [AWS](https://registry.terraform.io/providers/hashicorp/aws/latest/docs), [Azure](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs), [GCP](https://registry.terraform.io/providers/hashicorp/google/latest/docs), and any of the hundreds of other [Terraform providers](https://registry.terraform.io/browse/providers), to use them with your existing application configurations, through Terraform Cloud or Terraform Enterprise.
    * Deploy and Manage your Kubernetes and infrastructure resources in a single git repository, separate git repositories, or directly from a [module](https://www.terraform.io/docs/modules/index.html) in the [Terraform Registry](https://registry.terraform.io/browse/modules), to match your existing operating model.
    * Provide governance for your infrastructure resources using policy-as-code with [OPA Gatekeeper](https://github.com/open-policy-agent/gatekeeper) and [HashiCorp Sentinel](https://www.hashicorp.com/sentinel/).


You can read more about this project and its potential use cases on our [blog](https://www.hashicorp.com/blog/creating-workspaces-with-the-hashicorp-terraform-operator-for-kubernetes/).

Terraform also enables you to create and publish [custom infrastructure providers](https://learn.hashicorp.com/collections/terraform/providers) through the [Terraform SDK](https://www.terraform.io/docs/extend/plugin-sdk.html). Once you create a new Terraform provider, [publish it to the Terraform Registry](https://www.terraform.io/docs/registry/providers/publishing.html) and then you can consume it with the operator.

Join us in the [#terraform-providers channel on the Kubernetes Slack](https://kubernetes.slack.com/messages/CJY6ATQH4)  to discuss this, and other Terraform and Kubernetes projects ([Sign up here](http://slack.k8s.io/)). 

**Note:**
This project is versioned separately from Terraform. Supported Terraform versions must be version 0.12 or above. By versioning this project separately, we can iterate on Kubernetes integrations more quickly and release new versions without forcing Terraform users to do a full Terraform upgrade.

We take Terraform's security and our users' trust very seriously. If you believe you have found a security issue in the Terraform Cloud Operator for Kubernetes, please responsibly disclose by contacting us at security@hashicorp.com.

## Installation and Configuration

### Namespace

Create the namespace where you will deploy the operator, Secrets, and Workspace
resources.

```shell
$ kubectl create ns $NAMESPACE
```

### Authentication

The operator must authenticate to Terraform Cloud. Note that the operator must run within the cluster, which means that it already handles Kubernetes authentication.

1. Generate a Terraform Cloud Team API token at
   `https://app.terraform.io/app/$ORGANIZATION/settings/teams`, where
   `$ORGANIZATION` is your organization name.

1. Create a file for storing the API token and open it in a text editor.

1. Insert the generated token (`$TERRAFORM_CLOUD_API_TOKEN`) into the
   text file formatted for Terraform credentials.
   ```hcl
   credentials app.terraform.io {
     token = "$TERRAFORM_CLOUD_API_TOKEN"
   }
   ```

1. Create a Kubernetes secret named `terraformrc` in the namespace.
   Reference the credentials file (`$FILENAME`) created in the previous step.
   ```shell
   $ kubectl create -n $NAMESPACE secret generic terraformrc --from-file=credentials=$FILENAME
   ```
   Ensure `terraformrc` is the name of the secret, as it is the default secret name defined under the Helm value `syncWorkspace.terraformRC secretName` in the `values.yaml` file.

If you have the free tier of Terraform Cloud, you will only be able to generate a token for the one team associated with your account. If you have a paid tier of Terraform Cloud, create a separate team for the `operator` with "Manage Workspaces" access.

Note that a Terraform Cloud Team API token is a broad-spectrum token. It allows the token holder to create workspaces and execute Terraform runs. You cannot limit the access it provides to a single workspace or role within a team. In order to support a first-class Kubernetes experience, security and access control to this token must be enforced by [Kubernetes Role-Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) policies.

### Workspace Sensitive Variables

Sensitive variables in Terraform Cloud workspaces often take the form of credentials for cloud providers or API endpoints. They enable Terraform Cloud to authenticate against a provider and apply changes to infrastructure.

Create the secret for the namespace that contains all of the sensitive variables required for the workspace.

```shell
$ kubectl create -n $NAMESPACE secret generic workspacesecrets --from-literal=SECRET_KEY=$SECRET_KEY --from-literal=SECRET_KEY_2=$SECRET_KEY_2 ...
```
Ensure `workspacesecrets` is the name of the secret, as it is the default secret name defined under the Helm value `syncWorkspace.sensitiveVariables.secretName` in the `values.yaml` file.

In order to support a first-class Kubernetes experience, security and access control to these secrets must be enforced by [Kubernetes Role-Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) policies.

### Terraform Version

By default, the operator will create a Terraform Cloud workspace with
[a pinned version](https://github.com/hashicorp/terraform-k8s/blob/master/operator/version/version.go) of Terraform.

Override the Terraform version that will be set for the workspace by changing the Helm value `syncWorkspace.terraformVersion` to the Terraform version of choice.

## Deploy the Operator

Use the [Helm chart](https://github.com/hashicorp/terraform-helm) repository to deploy the Terraform Operator to the namespace you previously created.

```shell
$ helm repo add hashicorp https://helm.releases.hashicorp.com
$ helm search repo hashicorp/terraform
$ helm install --namespace ${NAMESPACE} hashicorp/terraform --generate-name
```

## Create a Workspace

The Workspace CustomResource defines a Terraform Cloud workspace, including
variables, Terraform module, and outputs.

Here are [examples of Workspace CustomResource](https://github.com/hashicorp/terraform-helm/tree/master/example)..

The Workspace Spec includes the following parameters:

1. `organization`: The Terraform Cloud organization you would like to use.

1. `secretsMountPath`: The file path defined on the operator deployment that contains the workspace's secrets.

Additional parameters are outlined below.

### Modules

> The Workspace will only execute Terraform configuration in a module. It will not execute `*.tf` files.

Information passed to the Workspace CustomResource will be rendered to a template Terraform configuration that uses the `module` block. Specify a module with remote `source`. Publicly available VCS repositories, the Terraform Registry, and private module registry are supported. In addition to `source`, specify a module `version`.

```yaml
module:
  source: "hashicorp/hello/random"
  version: "3.1.0"
```

The above Kubernetes definition renders to the following Terraform
configuration.

```hcl
module "operator" {
  source = "hashicorp/hello/random"
  version = "3.1.0"
}
```

### Variables

Variables for the workspace must equal the module's input variables.
You can define Terraform variables in two ways:

1. Inline
   ```yaml
   variables:
     - key: hello
       value: world
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

The above Kubernetes definition renders to the following Terraform
configuration.

```hcl
variable "hello" {}

variable "second_hello" {}

module "operator" {
  source = "hashicorp/hello/random"
  version = "3.1.0"
  hello = var.hello
  second_hello = var.second_hello
}
```

The operator pushes the values of the variables to the Terraform Cloud
workspace. For secrets, set `sensitive` to be `true`. The workspace sets them as write-only. Denote workspace environment variables by setting `environmentVariable` as `true`.

Sensitive variables should already be initialized as per [Workspace Sensitive Variables](#workspace-sensitive-variables). You can define them by setting `sensitive: true`. Do not define the value or use a ConfigMap reference, as the read from file will override the value you set.

```yaml
variables:
  - key: AWS_SECRET_ACCESS_KEY
    sensitive: true
    environmentVariable: true
```

### Apply an SSH key to the Workspace (optional)

SSH keys can be used to [clone private modules](https://www.terraform.io/docs/cloud/workspaces/ssh-keys.html). To apply an SSH key to the workspace, specify `sshKeyID` in the Workspace Custom Resource. The SSH key ID can be found in the [Terraform Cloud API](https://www.terraform.io/docs/cloud/api/ssh-keys.html#list-ssh-keys).

```
apiVersion: app.terraform.io/v1alpha1
kind: Workspace
metadata:
  name: $WORKSPACE
spec:
   sshKeyID: $SSHKEYID
```

### Outputs

In order to retrieve Terraform outputs, specify the `outputs`
section of the Workspace CustomResource. The `key` represents the output key you expect from `terraform output` and `moduleOutputName` denotes the module's output key name.

```yaml
outputs:
  - key: my_pet
    moduleOutputName: pet
```

The above Kubernetes definition renders to the following Terraform
configuration.

```hcl
output "my_pet" {
  value = module.operator.pet
}
```

The values of the outputs can be consumed from two places:

1. Kubernetes status of the workspace.
   ```shell
   $ kubectl describe -n $NAMESPACE workspace $WORKSPACE_NAME
   ```
1. ConfigMap labeled `$WORKSPACE_NAME-outputs`. Kubernetes deployments can consume these
   output values.
   ```shell
   $ kubectl describe -n $NAMESPACE configmap $WORKSPACE_NAME-outputs
   ```

### Deploy

Deploy the workspace after configuring its module, variables, and outputs.

```shell
$ kubectl apply -n $NAMESPACE -f workspace.yml
```

### Update a Workspace

The following changes updates and executes new runs for the Terraform Cloud workspace:

1. `organization`
1. `module` source or version
1. `outputs`
1. Non-sensitive or ConfigMap reference `variables`.

Updates to sensitive variables *will not* trigger a
new execution because sensitive variables are write-only for security
purposes. The operator is unable to reconcile the upstream value of the
secret with the value stored locally. Similarly, ConfigMap references do not trigger updates as the operator does not read the value for comparison.

After updating the configuration, re-deploy the workspace.

```shell
$ kubectl apply -n $NAMESPACE -f workspace.yml
```

### Delete a Workspace

In order for workspace destruction to work automatically, you must set the `CONFIRM_DESTROY` environment variable in the Terraform Cloud workspace. When you delete the Workspace CustomResource, the operator will attempt to destroy the workspace. As a secondary check, you must deploy the operator with this environment variable defined in the `variables` section if you would like to destroy the workspace in Terraform Cloud.

```yaml
variables:
  - key: CONFIRM_DESTROY
    value: "1"
    sensitive: false
    environmentVariable: true
```

When deleting the Workspace CustomResource, the command line will wait for a few moments.

```shell
$ kubectl delete -n $NAMESPACE workspace.app.terraform.io/$WORKSPACE_NAME
```

This is because the operator is running a [finalizer](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers). The finalizer will execute before the workspace officially deletes in order to:

1. Stop all runs in the workspace, including pending ones
1. `terraform destroy -auto-approve` on resources in the workspace
1. Delete the workspace.

Once the finalizer completes, Kubernetes deletes the Workspace CustomResource.

## Debugging

Check the status and outputs of the workspace by examining its Kubernetes status. This provides the run ID and workspace ID to debug in the Terraform Cloud UI.

```shell
$ kubectl describe -n $NAMESPACE workspace $WORKSPACE_NAME
```

When workspace creation, update, or deletion fails, check errors by
examining the logs of the operator.

```shell
$ kubectl logs -n $NAMESPACE $(kubectl get pods -n $NAMESPACE --selector "component=sync-workspace" -o jsonpath="{.items[0].metadata.name}")
```

If Terraform Cloud returns an error that the Terraform configuration is
incorrect, examine the Terraform configuration at its ConfigMap.

```shell
$ kubectl describe -n $NAMESPACE configmap $WORKSPACE_NAME
```

## Internals

### Why create a namespace and secrets?

The Helm chart does not include secrets management or injection. Instead, it expects to find secrets mounted as volumes to the operator's deployment. This supports secrets management approaches in Kubernetes that use a volume mount for secrets.

In order to support a first-class Kubernetes experience, security and access control to these secrets must be enforced by [Kubernetes Role-Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) policies.

For the Terraform Cloud Team API token, the entire credentials file with the Terraform Cloud API Token is mounted to the filepath specified by `TF_CLI_CONFIG_FILE`. In an equivalent Kubernetes configuration, the following example creates a Kubernetes secret and mount it to the operator at the filepath specified by `TF_CLI_CONFIG_FILE`.

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
      token = "$TERRAFORM_CLOUD_API_TOKEN"
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

Similar to the Terraform Cloud API Token, the Helm chart mounts
them to the operator's deployment for use. It __does not__ mount workspace sensitive variables to the Workspace Custom Resource. This ensures that only the operator has access to read and create sensitive variables as part of the Terraform Cloud workspace.

Examine the deployment in `templates/sync-workspace-deployment.yaml`. The deployment mounts a volume containing the sensitive variables. The file name is the secret's key and file contents is the secret's value. This supports secrets management approaches in Kubernetes that use a volume mount for secrets.

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

### Helm Chart

The Helm chart consists of several components. The Kubernetes configurations associated with the Helm chart are located under `crds/` and `templates/`.

#### Custom Resource Definition

Helm starts by deploying the Custom Resource Definition for the Workspace. Custom Resource Definitions extend the Kubernetes API. It looks for definitions in the `crds/` of the chart.

The Custom Resource Definition under `crds/app.terraform.io_workspaces_crd.yaml` defines that the Workspace Custom Resource schema.

#### Role-Based Access Control

In order to scope the operator to a namespace, Helm assigns a role and service account to the namespace. The role has access to Pods, Secrets, Services, and ConfigMaps. This configuration is located in `templates/`.

#### Namespace Scope

To ensure the operator does not have access to secrets or resource beyond the namespace, the Helm chart scopes the operator's deployment to a namespace.

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
          - /bin/terraform-k8s
          - "--k8s-watch-namespace=$(POD_NAMESPACE)"
          env:
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

When deploying, ensure that the namespace is passed into the
`--k8s-watch-namespace` option. Otherwise, the operator will attempt to access across all namespaces (cluster scope).
