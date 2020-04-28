## Unreleased

* Add `sshKeyID` to CR spec, so users can reference modules in private git repos (#25)
* Always update Sensitive variables when a Run is triggered and before the Run is executed (#22)

## 0.1.2 (April 16, 2020)

* Enable non-string Terraform variables by setting HCL type (#11)
* Fix panics in handling non-string Terraform output (#19)

## 0.1.1 (March 24, 2020)

* Handle non-string Terraform outputs and return them as JSON-formatted string (#12)

## 0.1.0 (2020)

* Initial release
