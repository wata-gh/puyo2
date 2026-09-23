FROM rust:1.94.0-bookworm AS runtime
# No keys, host mounts, Docker socket, or GitHub credentials in this image.
RUN apt-get update && apt-get install -y --no-install-recommends nodejs npm python3 zsh jq ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && npm install -g @openai/codex@0.157.0-alpha.11 \
    && codex exec-server --help >/dev/null
# jd is used by the repository's existing pnsolve comparison harness.
FROM golang:1.26.1-bookworm AS jd
RUN go install github.com/josephburnett/jd/v2/jd@v2.4.0
FROM runtime
COPY --from=jd /go/bin/jd /usr/local/bin/jd
RUN useradd --create-home --uid 10001 worker && mkdir /workspace /build && chown worker /workspace /build && chown -R worker /usr/local/cargo /usr/local/rustup
ENV CARGO_TARGET_DIR=/build CARGO_RESOLVER_INCOMPATIBLE_RUST_VERSIONS=fallback
USER worker
WORKDIR /workspace
CMD ["sleep", "infinity"]
