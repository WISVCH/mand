FROM golang AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install

FROM ghcr.io/wisvch/debian:bullseye-slim
WORKDIR /srv
COPY --from=builder /go/bin/mand /srv
COPY ./web /srv/web
USER 999
ENTRYPOINT ["/srv/mand"]
