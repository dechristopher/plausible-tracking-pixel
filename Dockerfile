# ---- Build Stage ----
FROM golang:1.23.4-alpine3.21 as builder

WORKDIR /build

ENV USER=svc
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Dependencies
COPY go.mod .
COPY go.sum .

# Pull and cache deps before code changes to improve cache hits on build
# RUN go mod download

# Copy vendored dependencies
COPY vendor vendor

# Embed static dependencies
COPY static static

# Copy service source
COPY main.go .
COPY internal internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main

# ---- Run Stage ----
FROM alpine:3.21 as prod

LABEL maintainer="Andrew DeChristopher"

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Import the user and group files from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Use an unprivileged user.
USER svc:svc

# Copy statically linked binary with embedded assets
COPY --from=builder build/main .

STOPSIGNAL SIGINT

ENTRYPOINT ["./main"]
