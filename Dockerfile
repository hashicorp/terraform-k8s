# Build the terraform-k8s binary
FROM golang:1.15-alpine as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o terraform-k8s main.go

FROM alpine:3.12.1
WORKDIR /
COPY --from=builder /workspace/terraform-k8s /bin/terraform-k8s
USER nobody:nobody

ENTRYPOINT ["/bin/terraform-k8s"]
