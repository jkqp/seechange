# One Dockerfile builds both binaries. APP picks which one:
#   docker build --build-arg APP=server .
#   docker build --build-arg APP=client .
ARG APP=server

# Build stage: compile a static binary.
FROM golang:1.26 AS build
ARG APP
WORKDIR /src

# Download dependencies first so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# zmq4 is pure Go, so cgo can be off and the binary needs no system libraries.
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/${APP}

# Run stage: just the binary on a minimal base image, running as non-root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
