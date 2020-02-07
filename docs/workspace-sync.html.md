---
layout: "docs"
page_title: "Workspace Sync - Terraform Cloud"
sidebar_current: "docs-platform-k8s-workspace-sync"
description: |-
  The workspaces in Kubernetes and Terraform Cloud can be automatically synced so that Kubernetes can execute runs and make updates to Terraform Cloud workspaces.
---

# Syncing Kubernetes and Terraform Cloud Workspaces

By creating the definition of a Workspace Custom Resource in Kubernetes,
workspace definitions in Kubernetes can be automatically synced to those in
Terraform Cloud by the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).
This functionality is provided by the
[terraform-k8s project](https://github.com/hashicorp/terraform-k8s) and can be
automatically installed and configured using the
[Terraform Helm chart](https://github.com/hashicorp/terraform-helm).

**Why create a Terraform Cloud Workspace Custom Resource Definition for Kubernetes?**
Applications deployed in Kubernetes will have the ability to define infrastructure
configuration using a Workspace in Kubernetes. The functionality depends on Terraform
Cloud to ensure consistent approaches to state locking, state storage, and execution.

**Why sync Kubernetes workspaces to Terraform Cloud?**
This functionality will re-execute updates to infrastructure and Terraform Cloud non-sensitive
variables. It provides a first-class Kubernetes interface to updating Terraform Cloud.

**How does it work?**
The workspace sync is done using a Terraform Operator in the
[terraform-k8s project](https://github.com/hashicorp/terraform-k8s). The Terraform Operator
must run within a Kubernetes cluster and be scoped to a namespace. The deployment
of the operator and resource definition is automated by a
[Helm chart](https://github.com/hashicorp/terraform-helm).

**How does the operator handle sensitive variables for Terraform Cloud?**
There are two categories of sensitive variables related to Terraform Cloud:
1. Terraform Cloud API Token: used to log in and execute runs for Terraform Cloud
2. Workspace Sensitive Variables: secrets that execution requires to log into providers (e.g., credentials)
See [Installation and Configuration](#InstallationandConfiguration) for how this is handled.

## Installation and Configuration

### Authentication

The operator must authenticate to Terraform Cloud. Note that `terraform-k8s` must
run within the cluster, which means already handles Kubernetes authentication. The Terraform Cloud API token can be generated under `https://app.terraform.io/app/${ORGANIZATION}/settings/teams`. Once generating this token, insert it into a file
formatted for Terraform credentials.

```hcl
credentials app.terraform.io {
  token = "${TERRAFORM_CLOUD_API_TOKEN}"
}
```

Note that the Terraform Cloud API token is a broad-spectrum token. It allows creation of workspaces and execution of runs. It will not offer access control on the workspace level or roles within the team. In order to support a first-class Kubernetes experience, security and access control to this token must be enforced by [Kubernetes Role-Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) policies.

As long as there is an environment variable called `TF_CLI_CONFIG_FILE` and the entire
credentials file with the Terraform Cloud API Token is mounted to the filepath specified
by `TF_CLI_CONFIG_FILE`, the operator will be able to authenticate against Terraform Cloud. This supports secrets management approaches in Kubernetes that use a volume mount for secrets.

For example, we create a Kubernetes secret and mount it to the operator at the filepath
specified by `TF_CLI_CONFIG_FILE`.

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
          command:
          - terraform-k8s
          - sync-workspace
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

Sensitive variables in Terraform Cloud workspaces often take the form of credentials
for cloud providers or API endpoints. They enable Terraform Cloud to authenticate
against a provider and apply changes to infrastructure.

Instead of defining sensitive variables in a Workspace CustomResource, you mount them to the operator's deployment for use. This ensures that only the operator has access to read and create sensitive variables as part of the Terraform Cloud workspace.
