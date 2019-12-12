# Terraform + Kubernetes (terraform-k8s)
This is a Kubernetes operator for [Terraform Cloud](https://app.terraform.io). Using this operator, you can define your Terraform Cloud organization and configuration along with your Kubernetes manifest files.

## Example Usage

1. Deploy roles and service accounts. `kubectl apply -n $NAMESPACE -f deploy/<file>.yaml`

1. Create a Kubernetes secret for the Terraform API token and workspace secrets.

1. Deploy the custom resource definition to our cluster for
   the workspace. `kubectl apply -n $NAMESPACE -f deploy/crds/app.terraform.io_workspaces_crd.yaml`

1. Create an `operator.yaml` that deploys the operator. It must mounts the Terraform API token
   and any workspace secrets we created in the previous step.
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: terraform-k8s
   spec:
     replicas: 1
     selector:
       matchLabels:
         name: terraform-k8s
     template:
       metadata:
         labels:
           name: terraform-k8s
       spec:
         serviceAccountName: terraform-k8s
         containers:
           - name: terraform-k8s
             image: joatmon08/operator-terraform
             command:
             - terraform-k8s
             imagePullPolicy: Always
             env:
               - name: WATCH_NAMESPACE
                 valueFrom:
                   fieldRef:
                     fieldPath: metadata.namespace
               - name: POD_NAME
                 valueFrom:
                   fieldRef:
                     fieldPath: metadata.name
               - name: OPERATOR_NAME
                 value: "terraform-k8s"
               - name: TF_CLI_CONFIG_FILE
                 value: "/etc/terraform/.terraformrc"
             volumeMounts:
             - name: terraformrc
               mountPath: "/etc/terraform"
               readOnly: true
             - name: workspacesecrets
               mountPath: "/tmp/secrets"
               readOnly: true
         volumes:
           - name: terraformrc
             secret:
               secretName: terraformrc
               items:
               - key: credentials
                 path: ".terraformrc"
           - name: workspacesecrets
             secret:
               secretName: workspace-secrets
   ```

1. Deploy the Workspace custom resource. Make sure the `secretsMountPath`
   points to the file path we used to mount workspace secrets.
   ```yaml
   apiVersion: app.terraform.io/v1alpha1
   kind: Workspace
   metadata:
     name: my-workspace
   spec:
     organization: rosemaryagain
     secretsMountPath: "/tmp/secrets"
     module:
       source: "app.terraform.io/rosemaryagain/hello/random"
       version: "2.0.1"
     variables:
       - key: hello
         value: rosemary
         sensitive: false
         environmentVariable: false
       - key: secret_key
         sensitive: true
         environmentVariable: false
       - key: AWS_SECRET_ACCESS_KEY
         sensitive: true
         environmentVariable: true
       - key: CONFIRM_DESTROY
         value: "1"
         sensitive: false
         environmentVariable: true
   ```