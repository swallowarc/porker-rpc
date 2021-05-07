# Build Container
FROM golang:1.16 as builder
WORKDIR /go/src/github.com/swallowarc/porker-rpc
COPY . .

# Set Environment Variable
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ARG NETRC_CONFIG

# Build
RUN echo $NETRC_CONFIG > $HOME/.netrc
RUN chmod 600 $HOME/.netrc
RUN make

# runtime image
FROM alpine:3.13.5
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/swallowarc/porker-rpc/bin /bin
ENTRYPOINT ["/bin/porker-rpc"]
