# Q4 2020 Roadmap

Each quarter the team will highlight areas of focus for our work and upcoming research.
 
We select items to include in the roadmap from community issues and internal priorities. When community pull requests exist for an item, we will prioritize working with the original authors to include their contributions. If the author can no longer take on the implementation, HashiCorp may complete any additional work needed. 

Each release will include necessary tasks that lead to the completion of the stated goals as well as community pull requests, enhancements, and features that are not highlighted in the roadmap. 

To make contribution easier, we’ll be using the [`Help Wanted`](https://github.com/hashicorp/terraform-k8s/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22) tag to point to issues we’d like to include in this quarter’s series of releases. Please review the [Contributing Guide](_about/CONTRIBUTING.md) for additional information.

This quarter (October-December ‘20) we will prioritize the following areas of work: 

## Currently In Progress

### General Availability (GA) 

Major version releases include code removals, deprecations, and breaking changes. A corresponding “upgrade guide” will be published alongside the release. 

In the GA release we'll focus on these areas:
 - [VCS Backed Workspaces](https://github.com/hashicorp/terraform-k8s/issues/59)
 - [Support for the latest version of Terraform](https://github.com/hashicorp/terraform-k8s/issues/64)
 - [Terraform Cloud Agents](https://github.com/hashicorp/terraform-k8s/issues/66)
 - [Sensitive Outputs as Secrets](https://github.com/hashicorp/terraform-k8s/issues/39)
 - [Upgrade to Operator SDK v1](https://github.com/hashicorp/terraform-k8s/issues/76)

#### Stretch Goals

These are items that aren't requirements for the operator to go GA, but are things we still want to address this quarter, if possible.

 - [Configure the Operator to deploy its CRD](https://github.com/hashicorp/terraform-k8s/issues/6)
 - [OperatorHub Listing](https://github.com/hashicorp/terraform-k8s/issues/57)
 - [TFC API Coverage](https://github.com/hashicorp/terraform-k8s/labels/theme%2Fcoverage)

## Feedback

We are interested in your thoughts and feedback about these proposals and encourage you to comment on the issue linked above or [schedule time with @redeux](https://calendly.com/philsautter/30min) to discuss.

## Disclosures

The product-development initiatives in this document reflect HashiCorp's current plans and are subject to change and/or cancellation in HashiCorp's sole discretion.