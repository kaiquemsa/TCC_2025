# ===== Etapa 1: Python base com o modelo =====
FROM python:3.11-slim AS python-base

# Evita problemas de I/O interativo
ENV DEBIAN_FRONTEND=noninteractive

# Define pasta de trabalho
WORKDIR /app

# Copia os scripts e dependências Python
COPY app/python /app/app/python

# Cria o ambiente virtual
RUN python -m venv /app/venv && \
    /app/venv/bin/pip install --no-cache-dir --upgrade pip && \
    /app/venv/bin/pip install --no-cache-dir -r /app/app/python/requirements.txt

# Copia o modelo já baixado localmente
COPY app/python/model /app/app/python/model

# Garante que os weights fiquem acessíveis
ENV TRANSFORMERS_OFFLINE=1
ENV HF_HUB_DISABLE_TELEMETRY=1

# ===== Etapa 2: Build do Go =====
FROM golang:1.23-bullseye AS builder

WORKDIR /app
COPY . .

# Resolve dependências e compila o binário
RUN go mod tidy && \
    go build -tags netgo -ldflags="-s -w" -o /app/server ./app/main.go

# ===== Etapa 3: Imagem final =====
FROM debian:bookworm-slim

# Pasta de trabalho
WORKDIR /app
ENV GOTOOLCHAIN=go1.23.0

# Copia o binário Go e o ambiente Python completo
COPY --from=builder /app/server /app/server
COPY --from=python-base /app/venv /app/venv
COPY --from=python-base /app/app/python /app/app/python

# Garante que o Python e o modelo sejam visíveis
ENV PATH="/app/venv/bin:$PATH"
ENV TRANSFORMERS_OFFLINE=1
ENV HF_HUB_DISABLE_TELEMETRY=1

# Executa o servidor Go
CMD ["/app/server"]