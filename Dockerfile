# syntax=docker/dockerfile:1

FROM golang:1.19-alpine AS builder
RUN apk --no-cache add git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go version
RUN go build --ldflags="-s -w" -o /server-tic-tac

FROM alpine
WORKDIR /app
COPY --from=builder /server-tic-tac /app
EXPOSE 9876
USER nonroot:nonroot
ENTRYPOINT ["/server-tic-tac"]