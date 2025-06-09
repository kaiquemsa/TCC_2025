package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type SupabaseService struct {
	url string
	key string
}

func NewSupabaseService(url, key string) *SupabaseService {
	return &SupabaseService{
		url: url,
		key: key,
	}
}

func (s *SupabaseService) Url() string {
	return s.url
}

func (s *SupabaseService) Key() string {
	return s.key
}

type QueryInput struct {
	SQL    string `json:"sql"`
	Salute bool   `json:"salute"`
	Response string `json:"response"`
}

func (s *SupabaseService) FindSimilarDocuments(embedding []float32) ([]map[string]interface{}, error) {
	if len(embedding) == 0 {
		return nil, fmt.Errorf("embedding vazio retornado")
	}

	var embeddingStr []string
	for _, value := range embedding {
		embeddingStr = append(embeddingStr, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	embeddingFormatted := "[" + strings.Join(embeddingStr, ",") + "]"

	body := map[string]interface{}{
		"query_embedding": embeddingFormatted,
		"match_threshold": 0.5,
		"match_count":     3,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", s.url+"/rest/v1/rpc/match_documents", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.key)
	req.Header.Set("Authorization", "Bearer "+s.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Erro na chamada RPC: %s", respBody)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *SupabaseService) ExecuteQuery(raw string) (interface{}, error) {
	cleaned := strings.Replace(raw, "```json", "", -1)
	cleaned = strings.Replace(cleaned, "```", "", -1)

	var input QueryInput
	err := json.Unmarshal([]byte(cleaned), &input)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer unmarshal: %v", err)
	}

	message := map[string]interface{}{
		"message": input.Response,
		"salute":  true,
	}	

	if input.Salute || input.SQL == "" {
		fmt.Println("Salute é true — ignorando execução da query.")
		return message, nil
	}

	query := input.SQL

	fmt.Println("Executando query:", query)

	body := map[string]interface{}{
		"query_string": query,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar o corpo da requisição: %v", err)
	}

	req, err := http.NewRequest("POST", s.url+"/rest/v1/rpc/execute_query", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar a requisição: %v", err)
	}

	req.Header.Set("apikey", s.key)
	req.Header.Set("Authorization", "Bearer "+s.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("erro ao enviar a requisição: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro na chamada RPC: %s, status: %d", respBody, resp.StatusCode)
	}

	var results []json.RawMessage
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("erro ao parsear JSON retornado: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("nenhum dado retornado da consulta")
	}

	var jsonData struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &jsonData); err != nil {
		var directData []map[string]interface{}
		if err := json.Unmarshal(respBody, &directData); err == nil {
			return directData, nil
		}
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return jsonData.Data, nil
}
