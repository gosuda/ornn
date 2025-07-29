FROM golang:1.24 AS builder
RUN apt-get update && apt-get install -y gcc-aarch64-linux-gnu

ADD . /workspace
WORKDIR /workspace
RUN CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux GOARCH=arm64 go build -o app .

FROM multiarch/ubuntu-core:arm64-bionic
WORKDIR /root/
COPY --from=builder /workspace .
CMD ["./ornn"]