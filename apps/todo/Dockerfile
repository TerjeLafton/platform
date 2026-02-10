FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/server ./apps/todo/cmd/server

FROM alpine:3.21
RUN adduser -D -h /app appuser
USER appuser
COPY --from=builder /bin/server /bin/server
ENTRYPOINT ["/bin/server"]
