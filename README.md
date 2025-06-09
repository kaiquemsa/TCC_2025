# Projeto TCC – Chat Inteligente com RAG e LLM

## Descrição

Este projeto é um **chat inteligente** integrado a um banco de dados PostgreSQL, utilizando os conceitos de **RAG (Retrieval-Augmented Generation)** e **LLM (Large Language Models)**. O objetivo é permitir que o usuário faça perguntas via chat e receba respostas personalizadas, precisas e baseadas nos dados armazenados no banco.

Ao receber uma consulta do usuário, o sistema busca os documentos mais semelhantes no banco utilizando **pgvector** (busca vetorial). Os resultados são enviados para a LLM Gemini, que analisa e gera a consulta SQL ideal para buscar a informação solicitada. O backend executa essa query, pega o resultado e, por fim, utiliza novamente a LLM para preparar o conteúdo de resposta apresentado ao usuário no frontend.

---

## Tecnologias Utilizadas

- **Backend:** Go (Golang) com framework [Fiber](https://gofiber.io/)
  - Integração com LLM Gemini (Google)
- **Frontend:** [Angular](https://angular.io/)
- **Banco de Dados:** [PostgreSQL](https://www.postgresql.org/)
  - Busca vetorial com [pgvector](https://github.com/pgvector/pgvector)
  - Gerenciamento com [Supabase](https://supabase.com/)

---

## Conceitos Principais

- **RAG (Retrieval-Augmented Generation):** Técnica para combinar busca de informações com geração de texto por IA, melhorando precisão e relevância das respostas.
- **LLM (Large Language Model):** Modelos de linguagem avançados, como o Gemini, utilizados para entender, processar e gerar texto de forma inteligente.

---

## Funcionalidades

- Chat em tempo real para consultas em linguagem natural
- Busca inteligente de documentos semelhantes usando vetores (pgvector)
- Geração automática e otimizada de consultas SQL via LLM
- Respostas personalizadas, formatadas pela IA, baseadas nos dados do banco
- Interface amigável construída em Angular

---

## Como rodar o projeto

### Pré-requisitos

- Go instalado ([download](https://go.dev/dl/))
- Node.js + npm ([download](https://nodejs.org/))
- PostgreSQL em funcionamento
- Conta/configuração do Supabase
- Chave de API para a Gemini LLM
- Clone o repositório

### Backend

1. Instale as dependências:
  ```bash
  cd tcc_2025/tcc_backend
  go mod tidy
  ```
2. Configure as variáveis de ambiente (.env) com as credenciais do banco, Supabase e Gemini.

3. Inicie o servidor:
  ```bash
  go run main.go
  ```
### Frontend

1. Instale as dependências:
  ```bash
  cd tcc_2025/tcc_frontend
  npm install
  ```
2. Inicie o servidor de desenvolvimento:
  ```bash
  ng serve
  ```
## Estrutura Simplificada do Fluxo

1. Usuário envia mensagem no chat.
2. Backend busca documentos semelhantes via pgvector (RAG).
3. Backend envia contexto para LLM Gemini gerar SQL.
4. Backend executa SQL no PostgreSQL.
5. Resultado volta para LLM Gemini preparar resposta.
6. Resposta formatada é enviada ao frontend.

## Observações
- O projeto exige tokens e configurações específicas de API (Gemini, Supabase etc.)

- Certifique-se de rodar as migrações do banco e configurar o pgvector.

- Para mais detalhes de implementação, veja a documentação interna de cada módulo.
