FROM rust:1.75-bookworm AS builder

RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY Cargo.toml ./
# Create a dummy main.rs to cache dependencies
RUN mkdir -p src && echo "fn main() {}" > src/main.rs
RUN cargo build --release 2>/dev/null || true
# Now copy real source
COPY src/ ./src/
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    bash \
    zsh \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/nucleus /usr/local/bin/nucleus

RUN mkdir -p /data/nucleus

ENV NUCLEUS_DATA_DIR=/data/nucleus
ENV RUST_LOG=info

ENTRYPOINT ["nucleus"]
