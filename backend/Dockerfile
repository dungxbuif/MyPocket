FROM golang:1.26-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.22
RUN adduser -D -H app
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY backend/migrations /app/migrations
USER app
ENTRYPOINT ["/app/api"]
