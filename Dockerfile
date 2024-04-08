FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/cernunnos cmd/main.go

EXPOSE 8080

CMD ["/app/cernunnos", "-log-level=debug", "-address=0.0.0.0:8080", "-db-host=cernunnos-db:5432", "-db-user=cernunnos", "-db-password=cernunnos"]
