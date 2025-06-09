package crud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

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
			id := row["id"]
			tabela := row["table_name"]
			coluna := row["column_name"]
			descricao := row["description"]

			text := fmt.Sprintf("Tabela %s, coluna %s: %s", tabela, coluna, descricao)

			embedding, err := handlers.GenerateEmbeddingHF(text)
			if err != nil {
				log.Println("Erro no embedding:", err)
				continue
			}

			// Passo 3: Atualizar o registro com PATCH
			payload := []map[string]interface{}{
				{
					"id":        id,
					"embedding": embedding,
				},
			}
			jsonPayload, _ := json.Marshal(payload)

			patchReq, _ := http.NewRequest("PATCH", supabaseService.Url()+"/rest/v1/struct", bytes.NewBuffer(jsonPayload))
			patchReq.Header.Set("apikey", supabaseService.Key())
			patchReq.Header.Set("Authorization", "Bearer "+supabaseService.Key())
			patchReq.Header.Set("Content-Type", "application/json")
			patchReq.Header.Set("Prefer", "resolution=merge-duplicates") // importante para PATCH em Supabase

			patchResp, err := client.Do(patchReq)
			if err != nil {
				log.Println("Erro no PATCH:", err)
				continue
			}
			defer patchResp.Body.Close()

			if patchResp.StatusCode >= 300 {
				respErr, _ := io.ReadAll(patchResp.Body)
				log.Printf("Erro ao atualizar ID %v: %s", id, string(respErr))
			}
		}

		return c.Status(200).JSON(fiber.Map{"message": "Embeddings atualizados com sucesso"})
	}
}
