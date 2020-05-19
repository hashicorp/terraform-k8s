## Unreleased

* Upgrade go-tfe to v0.7.0 and dependencies

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
