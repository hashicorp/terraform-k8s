crd:
	operator-sdk generate k8s
	operator-sdk generate openapi

docker:
	operator-sdk build joatmon08/operator-terraform
	docker push joatmon08/operator-terraform

operator:
	kubectl -n operator-test create secret generic terraformrc --from-file=./credentials || true
	kubectl -n operator-test create -f deploy/service_account.yaml || true
	kubectl -n operator-test create -f deploy/role.yaml || true
	kubectl -n operator-test create -f deploy/role_binding.yaml || true
	kubectl -n operator-test create -f deploy/crds/app.terraform.io_workspaces_crd.yaml || true
	kubectl -n operator-test create -f deploy/operator.yaml || true

workspace:
	kubectl -n operator-test create -f deploy/crds/app.terraform.io_v1alpha1_workspace_cr.yaml

clean:
	kubectl -n operator-test delete -f deploy/crds/app.terraform.io_v1alpha1_workspace_cr.yaml --ignore-not-found
	kubectl -n operator-test delete -f deploy/operator.yaml --ignore-not-found
	kubectl -n operator-test delete -f deploy/crds/app.terraform.io_workspaces_crd.yaml --ignore-not-found
	kubectl -n operator-test delete -f deploy/role_binding.yaml --ignore-not-found
	kubectl -n operator-test delete -f deploy/role.yaml --ignore-not-found
	kubectl -n operator-test delete -f deploy/service_account.yaml --ignore-not-found
	kubectl -n operator-test delete secret terraformrc --ignore-not-found