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


FROM scratch

WORKDIR /app
COPY --from=builder /app/cernunnos /app/cernunnos

CMD ["./cernunnos"]
