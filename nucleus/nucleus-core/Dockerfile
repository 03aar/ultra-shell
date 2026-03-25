FROM rust:1.75-bookworm AS builder

WORKDIR /app
COPY Cargo.toml ./
COPY src/ ./src/

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    bash \
    zsh \
    fish \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/nucleus /usr/local/bin/nucleus

RUN mkdir -p /data/nucleus

ENV NUCLEUS_DATA_DIR=/data/nucleus
ENV RUST_LOG=info

ENTRYPOINT ["nucleus"]
