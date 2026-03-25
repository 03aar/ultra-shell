FROM python:3.12-slim

LABEL maintainer="nucleus-orchestrator"
LABEL description="AI agent orchestration layer for the Nucleus shell engine"

WORKDIR /app

# System deps
RUN apt-get update && \
    apt-get install -y --no-install-recommends git && \
    rm -rf /var/lib/apt/lists/*

# Install Python package
COPY pyproject.toml .
COPY nucleus_orchestrator/ nucleus_orchestrator/

RUN pip install --no-cache-dir .

# Default env
ENV NUCLEUS_API_URL="http://host.docker.internal:8080/api/v1"
ENV NUCLEUS_API_KEY="dev-nucleus-key-local"

ENTRYPOINT ["nucleus-agent"]
CMD ["--help"]
