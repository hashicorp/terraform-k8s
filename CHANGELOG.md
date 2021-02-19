## 1.0.0 (February 19, 2021)

* Upgrade to operator SDK version 1.2.0 ([#83](https://github.com/hashicorp/terraform-k8s/pull/83))

## 0.2.1-beta (January 6, 2021)

* Revert state output parsing logic ([#86](https://github.com/hashicorp/terraform-k8s/pull/86))

## 0.2.0-beta (December 14, 2020)

Upgrade notes:
    This version moves storage of outputs from ConfigMaps to Secrets.
    The first time the operator runs it will create new Secrets containing
    the workspace's outputs, and it will keep updating only those for
    subsequent runs. Old ConfigMaps will be left for the users to delete
    when they are ready.

* Upgrade to operator-sdk v0.18 ([#69](https://github.com/hashicorp/terraform-k8s/pull/69))
* Support for VCS backed workspaces. ([#70](https://github.com/hashicorp/terraform-k8s/pull/70))
* Support for insecure HTTPS connections. ([#72](https://github.com/hashicorp/terraform-k8s/pull/72))
* Decouple the operator from the Terraform version used in a workspace ([#77](https://github.com/hashicorp/terraform-k8s/pull/77))
* Fix bug with timing of configuration version status ([#78](https://github.com/hashicorp/terraform-k8s/pull/78))
* Store outputs in k8s Secret objects instead of ConfigMap objects ([#80](https://github.com/hashicorp/terraform-k8s/pull/80))
* Remove vendor dir ([81](https://github.com/hashicorp/terraform-k8s/pull/81))
* Add support for TFC Remote agents and upgrade go-tfe to 0.11.1 ([#82](https://github.com/hashicorp/terraform-k8s/pull/82))


## 0.1.5-alpha (May 20, 2020)

* Upgrade go-tfe to v0.7.0 and dependencies 
* Fix issue that prevents SSHKey from being set on new workspaces (#44)
* Add Terraform Enterprise endpoint with `TF_URL` environment variable (#13)

## 0.1.4-alpha (May 14, 2020)

* Allow user to specify SSH Key by name or ID (#41)
* Update to Terraform 0.12.25 and add TF_VERSION environment variable (#47)

## 0.1.3-alpha (April 28, 2020)

* Add `sshKeyID` to CR spec, so users can reference modules in private git repos (#25)
* Always update Sensitive variables when a Run is triggered and before the Run is executed (#22)
* Fix: update variables when HCL flag changes (#33)

## 0.1.2-alpha (April 16, 2020)

* Enable non-string Terraform variables by setting HCL type (#11)
* Fix panics in handling non-string Terraform output (#19)

## 0.1.1-alpha (March 24, 2020)

* Handle non-string Terraform outputs and return them as JSON-formatted string (#12)

## 0.1.0-alpha (2020)

* Initial release
