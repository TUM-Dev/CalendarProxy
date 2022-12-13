FROM golang:1.19-alpine3.17 as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o /proxy cmd/proxy/proxy.go

FROM scratch

COPY --from=builder /proxy /proxy

EXPOSE 8080

CMD ["/proxy"]
