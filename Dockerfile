FROM golang:1-bullseye AS build
ENV CGO_ENABLED=1 \
    GOEXPERIMENT=boringcrypto \
    GOFLAGS='-trimpath "-ldflags=-s -w"'
WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.sum ./ 
RUN --mount=type=secret,id=gh_token,required=true \
  git config --global url."https://$(cat /run/secrets/gh_token):x-oauth-basic@github.com/snyk".insteadOf "https://github.com/snyk" && \
  go mod download

# Copy the go source
COPY . .
RUN go build -o terraform-k8s .

FROM gcr.io/snyk-main/ubuntu-20:2.1.0_202308221557
LABEL org.opencontainers.image.source=https://github.com/snyk/terraform-k8s
COPY --from=build /workspace/terraform-k8s /usr/local/bin/terraform-k8s
ENTRYPOINT ["/usr/local/bin/polaris-hello-world"]
