FROM golang:1.24.3-alpine AS builder

WORKDIR /service2

COPY ./service2/go.mod ./service2/go.sum ./
COPY ./service2/ .

RUN go mod download
RUN go build -o build/app cmd/app/main.go

FROM alpine:3.18

RUN mkdir -p /service2

COPY --from=builder ./service2/build/app /service2
COPY --from=builder ./service2/config.yaml /service2

CMD ["/service2/app"]
