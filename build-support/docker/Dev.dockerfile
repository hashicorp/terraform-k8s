FROM hashicorp/terraform:latest
COPY pkg/bin/linux_amd64/terraform-k8s /bin
ENTRYPOINT ["terraform-k8s"]