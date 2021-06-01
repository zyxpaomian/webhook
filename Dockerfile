FROM golang:1.16.2 as builder

RUN apt-get -y update && apt-get -y install upx

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY certs/ certs/
COPY common/ common/
COPY controller/ controller/
COPY http/ http/
RUN mkdir -p /config
COPY conf/sidecar.yaml /config/

# Build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download && \
    go build -a -o admission-registry main.go && \
    upx admission-registry

FROM alpine:3.9.2
COPY --from=builder /workspace/admission-registry .
ENTRYPOINT ["/admission-registry"]
