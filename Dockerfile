FROM golang:1.14 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /nss-bench main.go

FROM debian:latest
COPY --from=builder /nss-bench /nss-bench
CMD ["/nss-bench", "--help"]
