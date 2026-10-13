# syntax=docker/dockerfile:1

# The build stage runs natively and cross-compiles, so multi-platform builds
# (linux/amd64 + linux/arm64) do not need slow emulation.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# modernc.org/sqlite is pure Go, so the binary can be fully static.
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
COPY --from=build --chown=nonroot:nonroot /out/data /data

ENV ADDR=0.0.0.0:5000 \
    DB_PATH=/data/test.db
VOLUME ["/data"]
EXPOSE 5000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
