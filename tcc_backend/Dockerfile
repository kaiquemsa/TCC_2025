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

WORKDIR /app

# Instala Python e dependências do sistema
RUN apt-get update -y && \
    apt-get install -y --no-install-recommends python3 python3-venv python3-pip && \
    rm -rf /var/lib/apt/lists/*

# Copia o servidor Go
COPY --from=builder /app/server /app/server

# Copia o código Python (sem o venv prévio)
COPY --from=python-base /app/app/python /app/app/python

# Recria o venv na imagem final e instala requirements
RUN python3 -m venv /app/venv && \
    /app/venv/bin/pip install --no-cache-dir --upgrade pip && \
    /app/venv/bin/pip install --no-cache-dir -r /app/app/python/requirements.txt

# Define PATH e variáveis pra garantir execução offline
ENV PATH="/app/venv/bin:$PATH"
ENV TRANSFORMERS_OFFLINE=1
ENV HF_HUB_DISABLE_TELEMETRY=1

# Logs de verificação
RUN echo "Python path check:" && which python3 && python3 --version
RUN echo "Venv check:" && /app/venv/bin/python --version

# Executa o servidor Go
CMD ["/app/server"]