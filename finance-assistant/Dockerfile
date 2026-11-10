FROM python:3.13-slim AS builder
WORKDIR /build
COPY pyproject.toml ./
COPY src ./src
RUN python -m venv /venv && /venv/bin/pip install --no-cache-dir .

FROM python:3.13-slim
RUN useradd --system --uid 10001 --create-home assistant
COPY --from=builder /venv /venv
WORKDIR /app
RUN mkdir -p /data/workspaces && chown -R assistant:assistant /data /app
USER assistant
ENV PATH="/venv/bin:$PATH" PYTHONUNBUFFERED=1
EXPOSE 8090
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8090/healthz', timeout=3)"
CMD ["uvicorn", "finance_assistant.main:app", "--host", "0.0.0.0", "--port", "8090", "--proxy-headers", "--forwarded-allow-ips", "*"]
