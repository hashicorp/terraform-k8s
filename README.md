# Terraform Cloud + Kubernetes (terraform-k8s)

The `terraform-k8s` binary includes first-class integrations between Terraform and
Kubernetes. The project encapsulates multiple use cases, including a Terraform Cloud Operator
that synchronizes a Kubernetes Workspace (Custom Resource) to a Terraform Cloud Workspace.
This README will present a basic overview of each use case, but for full
documentation please reference the Terraform Cloud website.

This project is versioned separately from Terraform. Supported Terraform versions
must be above version 0.12. By versioning this project separately,
we can iterate on Kubernetes integrations more quickly and release new versions
without forcing Terraform users to do a full Terraform upgrade.

## Features

  * [**Terraform Cloud Workspace Sync**](docs/workspace-sync.html.md):
    Create and manage a Kubernetes Workspace that automatically synchronizes to Terraform Cloud.
    This enables Kubernetes to deploy infrastructure configured by Terraform.
    _(Requires Terraform 0.12+)_

## Installation

`terraform-k8s` is distributed in multiple forms:

  * The recommended installation method is the official
    [Terraform Helm chart](https://github.com/hashicorp/terraform-helm). This will
    automatically configure the Terraform and Kubernetes integration to run within
    an existing Kubernetes cluster.

  * A [Docker image `hashicorp/terraform-k8s`]() is available. This can be used to manually run `terraform-k8s` within a scheduled environment.

  * Raw binaries are available in the [HashiCorp releases directory]().
    These can be used to run `terraform-k8s` directly or build custom packages.
