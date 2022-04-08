FROM golang:1.16.4-buster AS builder

ARG VERSION=dev

WORKDIR /go/src/app
COPY main/main.go .
COPY main/go.mod .
RUN go mod tidy
RUN go build -o main -ldflags=-X=main.version=${VERSION} main.go

FROM debian:buster-slim
COPY --from=builder /go/src/app/main /go/bin/main
ENV PATH="/go/bin:${PATH}"
CMD ["main"]
