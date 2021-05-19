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
RUN mkdir -p /root/.ssh
RUN echo "$GITHUB_KEY" > /root/.ssh/id_rsa
RUN echo "StrictHostKeyChecking no" > /root/.ssh/config
RUN chmod 400 /root/.ssh/*
RUN git config --global url."git@github.com:swallowarc".insteadOf "https://github.com/swallowarc"
RUN make

# runtime image
FROM alpine:3.13.5
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/swallowarc/porker-rpc/bin /bin
ENTRYPOINT ["/bin/porker-rpc"]
