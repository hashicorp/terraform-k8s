#!/bin/bash

rm -rf pkg/client

client-gen \
	--input-dirs github.com/hashicorp/terraform-k8s/pkg/apis/app/v1alpha1/ \
	--input-base github.com/hashicorp/terraform-k8s/pkg/apis \
	--input app/v1alpha1 \
	-o pkg/ \
	--output-package github.com/hashicorp/terraform-k8s/pkg/client \
	--clientset-name clientset \
	--go-header-file hack/boilerplate.go.txt
mv pkg/github.com/hashicorp/terraform-k8s/pkg/client pkg/
rm -rf pkg/github.com
