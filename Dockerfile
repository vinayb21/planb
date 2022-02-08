FROM golang:1.17-alpine AS builder
RUN apk add git build-base
WORKDIR /go/src/planb
COPY go.mod .
RUN go mod tidy
RUN go mod download
COPY . .
RUN go build -o ./bin/planb .

FROM alpine:3.15
WORKDIR /planb
COPY --from=builder /go/src/planb/bin/planb ./
EXPOSE 8080
ENTRYPOINT ["./planb"]
