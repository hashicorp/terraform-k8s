# Terraform + Kubernetes (terraform-k8s)
This is a Kubernetes operator for [Terraform Cloud](https://app.terraform.io). Using this operator, you can define your Terraform Cloud organization and configuration along with your Kubernetes manifest files.

## Example Usage
```yaml
apiVersion: app.terraform.io/v1alpha1
kind: Organization
metadata:
  name: my-workspace
  namespace: my-tfc-org
spec:
  module:
    source: “terraform-aws-modules/security-group/aws”
    version: 3.2.0
  variables:
    - key: name
      value: example-security-group
    - key: vpc_id
      value: vpc-xxxxxxxxx
  environmentVariables:
    - sensitive: true
      key: AWS_SECRET_ACCESS_KEY
      value:
        - name: foo
          mountPath: "/etc/foo"
          readOnly: true
  volumes:
    - name: foo
      secret:
        secretName: mysecret

```