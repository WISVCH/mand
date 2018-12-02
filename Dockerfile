FROM golang:1.11-alpine AS build
# RUN apk add bash ca-certificates git gcc g++ libc-dev
RUN apk add ca-certificates git gcc g++
WORKDIR /go/src/github.com/wisvch/mand
ENV GO111MODULE=on
COPY . .
RUN go mod download
RUN go install

FROM alpine
COPY --from=build /go/bin/mand /bin/mand
COPY ./web ./web

ENTRYPOINT ["/bin/mand"]
