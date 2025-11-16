FROM golang:1.24.3-alpine AS builder

WORKDIR /service1

COPY ./service1/go.mod ./service1/go.sum ./
COPY ./service1/ .

RUN go mod download
RUN go build -o build/app cmd/app/main.go

FROM alpine:3.18

RUN mkdir -p /service1

COPY --from=builder ./service1/build/app /service1
COPY --from=builder ./service1/config.yaml /service1

EXPOSE 8081

CMD ["/service1/app"]
