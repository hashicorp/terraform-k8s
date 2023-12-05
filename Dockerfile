# Build the terraform-k8s binary
FROM golang:1.18-alpine as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=secret,id=gh_token,required=true \
  git config --global url."https://$(cat /run/secrets/gh_token):x-oauth-basic@github.com/snyk".insteadOf "https://github.com/snyk" && \
  go mod download

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o terraform-k8s main.go

FROM alpine:3.16.0
WORKDIR /
COPY --from=builder /workspace/terraform-k8s /bin/terraform-k8s
USER nobody:nobody

ENTRYPOINT ["/bin/terraform-k8s"]
