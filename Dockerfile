# syntax=docker/dockerfile:1

# --- Bazel build stage ---
FROM golang:1.25-bookworm AS bazel-base
# Install Bazel
RUN curl -fsSL https://github.com/bazelbuild/bazelisk/releases/download/v1.23.1/bazelisk-linux-amd64 -o /usr/local/bin/bazel \
    && chmod +x /usr/local/bin/bazel

WORKDIR /app
COPY . .

# Use Bazel to build (pure/static, no CGO needed)
RUN bazel build //cmd/server:server --platforms=@rules_go//go/toolchain:linux_amd64 \
    && bazel build //cmd/healthcheck:healthcheck --platforms=@rules_go//go/toolchain:linux_amd64

# --- Deployment stage ---
FROM cgr.dev/chainguard/static:latest

# Copy Bazel-built binaries
COPY --from=bazel-base /app/bazel-bin/cmd/server/server /weekly-rundown
COPY --from=bazel-base /app/bazel-bin/cmd/healthcheck/healthcheck /weekly-rundown-health

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  /weekly-rundown-health

ENTRYPOINT ["/weekly-rundown"]
