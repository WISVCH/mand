FROM golang AS builder
WORKDIR /go/src/github.com/wisvch/mand
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install

FROM wisvch/debian:stretch-slim
WORKDIR /srv
COPY --from=builder /go/bin/mand /srv
COPY ./web /srv/web

RUN groupadd -r mand --gid=999 && useradd --no-log-init -r -g mand --uid=999 mand
USER 999
ENTRYPOINT ["/srv/mand"]
