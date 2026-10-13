# syntax=docker/dockerfile:1

FROM python:3.13-slim

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PYTHONPATH=/app

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY api.py utils.py openapi.yaml docs.html ./

RUN useradd --system --uid 10001 app \
    && mkdir /data && chown app:app /data
USER app

# api.py creates test.db in the working directory.
WORKDIR /data
VOLUME ["/data"]
EXPOSE 5000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:5000/active', timeout=2)"

CMD ["waitress-serve", "--host=0.0.0.0", "--port=5000", "api:app"]
