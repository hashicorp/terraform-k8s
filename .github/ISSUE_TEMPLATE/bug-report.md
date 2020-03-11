---
name: "\U0001F41B Bug Report"
about: "If something isn't working as expected \U0001F914."
title: ''
labels: bug
assignees: ''

---

<!---
Please note the following potential times when an issue might be in Terraform core:

* [Configuration Language](https://www.terraform.io/docs/configuration/index.html) or resource ordering issues
* [State](https://www.terraform.io/docs/state/index.html) and [State Backend](https://www.terraform.io/docs/backends/index.html) issues
* [Provisioner](https://www.terraform.io/docs/provisioners/index.html) issues
* [Registry](https://registry.terraform.io/) issues
* Spans resources across multiple providers

If you are running into one of these scenarios, we recommend opening an issue in the [Terraform core repository](https://github.com/hashicorp/terraform/) instead.
--->

<!--- Please keep this note for the community --->

### Community Note

* Please vote on this issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment

<!--- Thank you for keeping this note for the community --->

### terraform-k8s & Kubernetes Version

<!--- Please make note of the image version you are using for the `hashicorp/terraform-k8s` container and the version of Kubernetes you are using.
If you are not running the latest version of the operator, please upgrade because your issue may have already been fixed. --->

### Affected Resource(s)

<!--- Please list the affected resources and data sources. --->

### Terraform Configuration Files

<!--- Information about code formatting: https://help.github.com/articles/basic-writing-and-formatting-syntax/#quoting-code --->

```hcl
# Copy-paste your Terraform configuration from the operator here.
# To retrieve the configuration, use `kubectl -n $NAMESPACE describe configmap $WORKSPACE_NAME`
```

### Debug Output

<!---
Please provide a link to a GitHub Gist containing the complete debug output. Please do NOT paste the debug output in the issue; just paste a link to the Gist.

To retrieve debug information, run ` kubectl -n $NAMESPACE logs $(kubectl -n $NAMESPACE get pods --selector "component=sync-workspace" -o jsonpath="{.items[0].metadata.name}")`.
--->

### Expected Behavior

<!--- What should have happened? --->

### Actual Behavior

<!--- What actually happened? --->

### Steps to Reproduce

<!--- Please list the steps required to reproduce the issue. --->

### Important Factoids

<!--- Are there anything atypical about your accounts that we should know? For example: Kubernetes RBAC? --->

### References

<!---
Information about referencing Github Issues: https://help.github.com/articles/basic-writing-and-formatting-syntax/#referencing-issues-and-pull-requests

Are there any other GitHub issues (open or closed) or pull requests that should be linked here? Vendor documentation? For example:
--->

* #0000
