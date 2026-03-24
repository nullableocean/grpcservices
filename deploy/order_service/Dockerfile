FROM golang:1.25-alpine AS builder

WORKDIR /app/build

# go.mod --> replace modules api, shared as ../api, ../shared
COPY api/ /app/api
COPY shared/ /app/shared

COPY order_service/go.mod order_service/go.sum ./
RUN go mod download

COPY order_service/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --chown=appuser:appgroup --from=builder /app/build/app .

USER appuser

CMD ["./app"]