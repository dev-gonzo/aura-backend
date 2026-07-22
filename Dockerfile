FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/editora-backend .

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache wget ca-certificates
RUN adduser -D -g '' appuser

COPY --from=builder /bin/editora-backend /usr/local/bin/editora-backend

USER appuser

EXPOSE 8081

ENTRYPOINT ["/usr/local/bin/editora-backend"]
