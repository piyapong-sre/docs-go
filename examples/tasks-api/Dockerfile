# syntax=docker/dockerfile:1

# ---- Build stage: compile a static binary ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Download modules first so this layer is cached when only code changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /out/api ./cmd/api

# ---- Runtime stage: no shell, no package manager, non-root user ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
