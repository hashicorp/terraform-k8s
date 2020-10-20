# Contributing

To build and install `terraform-k8s` locally, Go version 1.12.13+ is required.
You will also need to install the Docker engine:

- [Docker for Mac](https://docs.docker.com/engine/installation/mac/)
- [Docker for Windows](https://docs.docker.com/engine/installation/windows/)
- [Docker for Linux](https://docs.docker.com/engine/installation/linux/ubuntulinux/)

Clone the repository:

```shell
$ git clone https://github.com/hashicorp/terraform-k8s.git
```

To compile the `terraform-k8s` binary for your local machine:

```shell
$ make dev
```

This will compile the `terraform-k8s` binary into `bin/terraform-k8s` as
well as your `$GOPATH` and run the test suite.

Or run the following to generate all binaries:

```shell
$ make dist
```

If you just want to run the tests:

```shell
$ make test
```

Or to run a specific test in the suite:

```shell
go test ./... -run SomeTestFunction_name
```

Example of running specific acceptance tests against Terraform Cloud and Terraform Enterprise:
```shell
export TF_URL="https://my-tfe-hostname"
export TF_ACC=1
export TF_CLI_CONFIG_FILE=$HOME/.terraformrc
export TF_ORG="my-tfe-org"
go test -v ./... -run TestOrganizationTerraformEnterprise

export TF_ORG="my-tfc-org"
go test -v ./... -run TestOrganizationTerraformCloud
```

To create a docker image with your local changes:

```shell
$ make dev-docker
```

The operator under `terraform-k8s sync-workspace` uses the 
[Operator SDK](https://github.com/operator-framework/operator-sdk) to manage
and generate custom resource definitions. To generate new OpenAPI specifications
and CustomResource Definitions, go into the operator's Makefile.

```shell
$ cd operator
$ make crd
```

### Rebasing contributions against master

PRs in this repo are merged using the [`rebase`](https://git-scm.com/docs/git-rebase) method. This keeps
the git history clean by adding the PR commits to the most recent end of the commit history. It also has
the benefit of keeping all the relevant commits for a given PR together, rather than spread throughout the
git history based on when the commits were first created.

If the changes in your PR do not conflict with any of the existing code in the project, then Github supports
automatic rebasing when the PR is accepted into the code. However, if there are conflicts (there will be
a warning on the PR that reads "This branch cannot be rebased due to conflicts"), you will need to manually
rebase the branch on master, fixing any conflicts along the way before the code can be merged.
