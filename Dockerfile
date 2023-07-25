FROM golang:bookworm AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install

FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /srv
COPY --from=builder /go/bin/mand /srv
COPY ./web /srv/web
USER 999
ENTRYPOINT ["/srv/mand"]
