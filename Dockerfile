FROM golang:1.21-alpine as builder
# from makefile
ARG VERSION

RUN mkdir -p /linode
WORKDIR /linode

ARG VERSION

COPY go.mod .
COPY go.sum .
COPY main.go .
#COPY pkg ./pkg

RUN go mod download
RUN GOARCH=amd64 go build -a -ldflags '-X main.version='${VERSION}' -extldflags "-static"' -o /bin/k8s-node-decorator /linode

FROM alpine:3.18.5
LABEL maintainers="Linode"
LABEL description="Linode Kubernetes Node Decorator"

COPY --from=builder /bin/k8s-node-decorator /k8s-node-decorator

ENTRYPOINT ["/k8s-node-decorator"]
