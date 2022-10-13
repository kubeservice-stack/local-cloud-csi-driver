# Build the manager binary
FROM golang:1.18.5-alpine as builder

RUN apk add --no-cache gcc musl-dev libc6-compat

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/
COPY hack/ hack/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-linkmode external -extldflags -static" -o local-cloud-csi-driver ./main.go

FROM ubuntu:20.04

RUN apt update && apt upgrade -y && apt install -y ca-certificates file tzdata lvm2

COPY --from=builder /workspace/local-cloud-csi-driver /bin/local-cloud-csi-driver
COPY --from=builder /workspace/hack/local/entrypoint.sh /entrypoint.sh
COPY --from=builder /workspace/hack/local/nsenter /nsenter

RUN chmod +x /bin/local-cloud-csi-driver && chmod +x /entrypoint.sh && sed -i 's/use_lvmetad\ =\ 1/use_lvmetad\ =\ 0/g' /etc/lvm/lvm.conf

ENTRYPOINT ["/entrypoint.sh"]
