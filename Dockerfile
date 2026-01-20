FROM golang:1.21-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /tpsp ./cmd/tpsp


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/nsswitch.conf /etc/nsswitch.conf
COPY --from=builder /tpsp /tpsp

ENTRYPOINT ["/tpsp"]
