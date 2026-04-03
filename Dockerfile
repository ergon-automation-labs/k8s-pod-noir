# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm AS builder
WORKDIR /src
ARG VERSION=dev
ARG COMMIT=none
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -X podnoir/internal/version.Version=${VERSION} -X podnoir/internal/version.Commit=${COMMIT}" \
    -o /out/podnoir ./cmd/podnoir

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates curl \
	&& ARCH=$(uname -m) \
	&& case "$ARCH" in \
		x86_64) KARCH=amd64 ;; \
		aarch64) KARCH=arm64 ;; \
		*) echo "unsupported arch: $ARCH" && exit 1 ;; \
	esac \
	&& curl -fsSL -o /tmp/kubectl "https://dl.k8s.io/release/v1.29.13/bin/linux/${KARCH}/kubectl" \
	&& install -m 0755 /tmp/kubectl /usr/local/bin/kubectl \
	&& rm /tmp/kubectl
COPY --from=builder /out/podnoir /usr/local/bin/podnoir
ENTRYPOINT ["/usr/local/bin/podnoir"]

FROM golang:1.23-bookworm AS dev
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
CMD ["go", "test", "./..."]
