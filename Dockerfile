FROM golang:1.22.4 AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y \
    curl xz-utils \
    gcc g++ mingw-w64 \
    gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
    cmake libssl-dev libxml2-dev vim apt-transport-https \
    zip unzip libtinfo5 patch zlib1g-dev autoconf libtool \
    pkg-config make docker.io gnupg2 libgmp-dev python3

ENV GOPROXY=https://gomodules.cbhq.net/
COPY ./go.mod ./go.sum /app/

RUN go mod download

COPY . /app/

# The flag below may be needed if blst throws SIGILL, which happens with certain (older) CPUs
# ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"

RUN go build -o /build/beacon-chain -buildvcs=false ./cmd/beacon-chain
RUN go build -o /build/validator -buildvcs=false ./cmd/validator
RUN go build -o /build/bootnode -buildvcs=false ./tools/bootnode

FROM ubuntu:24.04
RUN apt-get update && apt-get install -y curl jq iproute2

COPY --from=builder /build/beacon-chain /beacon-chain
# COPY --from=builder /build/validator /usr/local/bin/

# WORKDIR /usr/local/bin
ENTRYPOINT [ "/beacon-chain" ]