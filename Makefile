NAMESPACE='rosemarycorp'

test:
	OPERATOR_NAME=terraform-k8s operator-sdk up local --namespace=$(NAMESPACE)

crd:
	operator-sdk generate k8s
	operator-sdk generate openapi

docker:
	operator-sdk build joatmon08/operator-terraform
	docker push joatmon08/operator-terraform

setup:
	kubectl create ns $(NAMESPACE) || true
	kubectl -n $(NAMESPACE) create secret generic terraformrc --from-file=./credentials || true
	kubectl -n $(NAMESPACE) create -f deploy/service_account.yaml || true
	kubectl -n $(NAMESPACE) create -f deploy/role.yaml || true
	kubectl -n $(NAMESPACE) create -f deploy/role_binding.yaml || true
	kubectl -n $(NAMESPACE) create -f deploy/crds/app.terraform.io_workspaces_crd.yaml || true

operator: docker setup
	kubectl -n $(NAMESPACE) create -f deploy/operator.yaml || true

workspace:
	kubectl -n $(NAMESPACE) create -f deploy/crds/app.terraform.io_v1alpha1_workspace_cr.yaml

clean-workspace:
	kubectl -n $(NAMESPACE) delete -f deploy/crds/app.terraform.io_v1alpha1_workspace_cr.yaml --ignore-not-found

clean: clean-workspace
	kubectl -n $(NAMESPACE) delete -f deploy/operator.yaml --ignore-not-found
	kubectl -n $(NAMESPACE) delete -f deploy/crds/app.terraform.io_workspaces_crd.yaml --ignore-not-found
	kubectl -n $(NAMESPACE) delete -f deploy/role_binding.yaml --ignore-not-found
	kubectl -n $(NAMESPACE) delete -f deploy/role.yaml --ignore-not-found
	kubectl -n $(NAMESPACE) delete -f deploy/service_account.yaml --ignore-not-found
	kubectl -n $(NAMESPACE) delete secret terraformrc --ignore-not-found
	kubectl delete ns $(NAMESPACE)