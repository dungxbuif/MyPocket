FROM node:22-alpine AS docs-build
WORKDIR /app/docs
COPY frontend/docs/package.json frontend/docs/package-lock.json ./
RUN npm ci
COPY frontend/docs/ .
RUN npm run build

FROM golang:1.24-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.21
RUN adduser -D -H app
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker
COPY --from=build /out/migrate /app/migrate
COPY backend/migrations /app/migrations
COPY --from=docs-build /app/docs/build /app/docs
USER app
ENTRYPOINT ["/app/api"]