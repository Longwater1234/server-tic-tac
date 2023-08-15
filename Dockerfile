# syntax=docker/dockerfile:1

FROM golang:alpine3.18 AS builder
RUN apk --no-cache add git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go version
RUN go build --ldflags="-s -w" -o /server-tic-tac

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /server-tic-tac /app
EXPOSE 9876
USER nonroot:nonroot
ENTRYPOINT ["/server-tic-tac"]