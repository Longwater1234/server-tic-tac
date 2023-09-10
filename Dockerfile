# syntax=docker/dockerfile:1

FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./

# IF LOCATED IN MAINLAND CHINA, UNCOMMENT THESE 2 LINES BELOW
#RUN go env -w GO111MODULE=on
#RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod download
COPY . ./
RUN go version
RUN go build --ldflags="-s -w" -o server-tic-tac

FROM alpine
WORKDIR /app
COPY --from=builder /app/server-tic-tac /app
EXPOSE 9876
ENTRYPOINT ["./server-tic-tac"]