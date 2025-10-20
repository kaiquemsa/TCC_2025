package crud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kaiquemsa/nlp-sql-backend/app/handlers"
	"github.com/kaiquemsa/nlp-sql-backend/app/internal/supabase"
)

func SaveChatMessage(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var chatMessages []Chat
		if err := c.BodyParser(&chatMessages); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid data", "details": err.Error()})
		}

		for _, chat := range chatMessages {
			jsonBody, err := json.Marshal(chat)
			if err != nil {
				log.Println("Erro ao serializar mensagem:", err)
				continue
			}

			req, err := http.NewRequest("POST", supabaseService.Url()+"/rest/v1/chat_history", bytes.NewBuffer(jsonBody))
			if err != nil {
				log.Println("Erro ao criar request:", err)
				continue
			}

			req.Header.Set("apikey", supabaseService.Key())
			req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Prefer", "return=representation")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Erro ao enviar request:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 300 {
				respBody, _ := io.ReadAll(resp.Body)
				log.Println("Erro ao salvar no Supabase:", string(respBody))
				continue
			}
		}

		return c.Status(201).JSON(fiber.Map{"message": "Todas mensagens processadas com sucesso"})
	}
}

func GetHistory(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uuid := c.Query("uuid")
		top := c.Query("top")

		if uuid == "" {
			return c.Status(400).JSON(fiber.Map{"error": "uuid não informado"})
		}

		body, err := FetchHistory(uuid, supabaseService, top)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar histórico"})
		}

		return c.Status(200).Send(body)
	}
}

func FetchHistory(uuid string, supabaseService *supabase.SupabaseService, top string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/chat_history?id_chat=eq.%s", supabaseService.Url(), uuid)
	if top != "" {
		url += fmt.Sprintf("&limit=%s", top)
	}
	// Adiciona o order by DESC sempre
	url += "&order=created_at.desc"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", supabaseService.Key())
	req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

type RecentChat struct {
	IDChat        string    `json:"id_chat"`
	LastMessageAt time.Time `json:"last_message_at"`
}

func GetRecentChats(supabase *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := 10
		if top := c.Query("top"); top != "" {
			if n, err := strconv.Atoi(top); err == nil && n > 0 {
				limit = n
			}
		}

		v := url.Values{}
		v.Set("order", "last_message_at.desc")
		v.Set("limit", strconv.Itoa(limit))

		endpoint := fmt.Sprintf("%s/rest/v1/chat_recent?%s", supabase.Url(), v.Encode())

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		req.Header.Set("apikey", supabase.Key())
		req.Header.Set("Authorization", "Bearer "+supabase.Key())
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(resp.Body)
			return c.Status(resp.StatusCode).Send(b)
		}

		var items []RecentChat
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(items)
	}
}

func GetEmbeddings(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var embeddings []map[string]interface{}

		req, err := http.NewRequest("GET", supabaseService.Url()+"/rest/v1/struct?embedding=is.null", nil)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get request", "details": err.Error()})
		}

		req.Header.Set("apikey", supabaseService.Key())
		req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to send request", "details": err.Error()})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read response", "details": err.Error()})
		}

		if resp.StatusCode >= 300 {
			return c.Status(resp.StatusCode).JSON(fiber.Map{
				"error":   "Failed to retrieve embeddings",
				"details": string(respBody),
			})
		}

		if err := json.Unmarshal(respBody, &embeddings); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse response", "details": err.Error()})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Embeddings retrieved successfully", "data": embeddings})
	}
}

func UpdateEmbedding(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing 'id' parameter"})
		}

		var embedding Embedding
		if err := c.BodyParser(&embedding); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid data"})
		}

		jsonBody, err := json.Marshal(embedding)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to serialize embedding", "details": err.Error()})
		}

		req, err := http.NewRequest("PATCH", supabaseService.Url()+"/rest/v1/documents?id=eq."+id, bytes.NewBuffer(jsonBody))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create request", "details": err.Error()})
		}

		req.Header.Set("apikey", supabaseService.Key())
		req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=representation")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to send request", "details": err.Error()})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read response", "details": err.Error()})
		}

		if resp.StatusCode >= 300 {
			return c.Status(resp.StatusCode).JSON(fiber.Map{
				"error":   "Failed to update embedding",
				"details": string(respBody),
			})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Embedding updated successfully"})
	}
}

func DeleteEmbedding(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing 'id' parameter"})
		}

		req, err := http.NewRequest("DELETE", supabaseService.Url()+"/rest/v1/documents?id=eq."+id, nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create request", "details": err.Error()})
		}

		req.Header.Set("apikey", supabaseService.Key())
		req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to send request", "details": err.Error()})
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read response", "details": err.Error()})
		}

		if resp.StatusCode >= 300 {
			return c.Status(resp.StatusCode).JSON(fiber.Map{
				"error":   "Failed to delete embedding",
				"details": string(respBody),
			})
		}

		return c.Status(200).JSON(fiber.Map{"message": "Embedding deleted successfully"})
	}
}

func GenerateEmbeddingsFromStruct(supabaseService *supabase.SupabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []map[string]interface{}

		// Passo 1: Buscar registros da tabela `struct` onde o embedding ainda é nulo
		req, err := http.NewRequest("GET", supabaseService.Url()+"/rest/v1/struct?embedding=is.null", nil)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Erro na requisição", "details": err.Error()})
		}

		req.Header.Set("apikey", supabaseService.Key())
		req.Header.Set("Authorization", "Bearer "+supabaseService.Key())
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao buscar registros", "details": err.Error()})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &rows); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Erro ao processar resposta", "details": err.Error()})
		}

		if len(rows) == 0 {
			return c.Status(200).JSON(fiber.Map{"message": "Nenhum registro para processar"})
		}

		// Passo 2: Iterar sobre os registros e gerar embeddings
		for _, row := range rows {
			// Extrair campos
			idVal, ok := row["id"]
			if !ok {
				log.Println("Registro sem ID, ignorando...")
				continue
			}

			// Corrigir tipo do ID (pode vir como float64 do JSON)
			var id string
			switch v := idVal.(type) {
			case float64:
				id = strconv.FormatInt(int64(v), 10)
			case string:
				id = v
			default:
				log.Printf("Tipo inesperado para ID: %T\n", v)
				continue
			}

			tabela := row["table_name"]
			coluna := row["column_name"]
			descricao := row["description"]

			text := fmt.Sprintf("Tabela %v, coluna %v: %v", tabela, coluna, descricao)

			embedding, err := handlers.GenerateEmbeddingLocal(text)
			if err != nil {
				log.Println("Erro ao gerar embedding:", err)
				continue
			}

			// Passo 3: Atualizar o registro com PATCH (usando WHERE id)
			url := fmt.Sprintf("%s/rest/v1/struct?id=eq.%s", supabaseService.Url(), id)

			payload := map[string]interface{}{
				"embedding": embedding,
			}
			jsonPayload, _ := json.Marshal(payload)

			patchReq, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonPayload))
			patchReq.Header.Set("apikey", supabaseService.Key())
			patchReq.Header.Set("Authorization", "Bearer "+supabaseService.Key())
			patchReq.Header.Set("Content-Type", "application/json")

			patchResp, err := client.Do(patchReq)
			if err != nil {
				log.Printf("Erro no PATCH para ID %s: %v\n", id, err)
				continue
			}
			defer patchResp.Body.Close()

			if patchResp.StatusCode >= 300 {
				respErr, _ := io.ReadAll(patchResp.Body)
				log.Printf("Erro ao atualizar ID %s: %s\n", id, string(respErr))
			} else {
				log.Printf("Embedding atualizado com sucesso para ID %s\n", id)
			}
		}

		return c.Status(200).JSON(fiber.Map{"message": "Embeddings atualizados com sucesso"})
	}
}
