# Build Container
FROM golang:1.16 as builder
WORKDIR /go/src/github.com/swallowarc/porker-rpc
COPY . .

# Set Environment Variable
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ARG GITHUB_KEY

# Build
RUN make

# runtime image
FROM alpine:3.13.5
RUN apk --no-cache add ca-certificates
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.4 && \
  wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
  chmod +x /bin/grpc_health_probe
COPY --from=builder /go/src/github.com/swallowarc/porker-rpc/bin /bin
ENTRYPOINT ["/bin/porker-rpc"]
