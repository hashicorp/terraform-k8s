# Terraform + Kubernetes (terraform-k8s)

> This experimental repository contains software which is still being developed
> and in the alpha testing stage. It is not ready for production use.

The `terraform-k8s` binary includes first-class integrations between Terraform
and Kubernetes. Currently, this project only includes the Terraform Cloud
Operator, which synchronizes a Kubernetes Workspace (Custom Resource) to a
Terraform Cloud Workspace.You can read more about this project and its potential use cases on our [blog](https://www.hashicorp.com/blog/creating-workspaces-with-the-hashicorp-terraform-operator-for-kubernetes/).
We are actively considering other possible use cases to add to this project outside of the operator, and welcome your feedback. 

This project is versioned separately from Terraform. Supported Terraform
versions must be above version 0.12. By versioning this project separately, we
can iterate on Kubernetes integrations more quickly and release new versions
without forcing Terraform users to do a full Terraform upgrade.

## Features

  * [**Terraform Cloud Workspace Sync**](https://github.com/hashicorp/terraform-helm/blob/master/docs/workspace-sync.html.md): Create
    and manage a Kubernetes Workspace that automatically synchronizes to
    Terraform Cloud. This enables Kubernetes to deploy infrastructure configured
    by Terraform. _(Requires Terraform 0.12+)_

## Installation

`terraform-k8s` is distributed in multiple forms:

  * The recommended installation method is the official [Terraform Helm
    chart](https://github.com/hashicorp/terraform-helm). This will automatically
    configure the Terraform and Kubernetes integration to run within an existing
    Kubernetes cluster.

  * A [Docker image
    `hashicorp/terraform-k8s`](https://hub.docker.com/repository/docker/hashicorp/terraform-k8s)
    is available. This can be used to manually run `terraform-k8s` within a
    scheduled environment.
