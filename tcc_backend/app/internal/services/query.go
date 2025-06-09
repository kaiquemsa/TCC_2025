package services

import (
	"fmt"
	"strings"
	"encoding/json"

	"github.com/kaiquemsa/nlp-sql-backend/app/internal/supabase"
	"github.com/kaiquemsa/nlp-sql-backend/app/crud"
)

type QueryService struct {
	gemini   *GeminiService
	supabase *supabase.SupabaseService
}

func NewQueryService(gemini *GeminiService, supabase *supabase.SupabaseService) *QueryService {
	return &QueryService{
		gemini:   gemini,
		supabase: supabase,
	}
}

func (s *QueryService) ProcessQuery(question string, history string, uuid string) (string, interface{}, error) {
	fmt.Println("Args iniciais:", question, history, uuid)	
	// Caso haja historico
	var formattedHistory []string
	if history == "Y" {
		var top = "6"
		chatHistory, err := crud.FetchHistory(uuid, s.supabase, top)
		if err != nil {
			return "", nil, fmt.Errorf("erro ao buscar histórico: %v", err)
		}
	
		fmt.Println(string(chatHistory))
	
		var messages []struct {
			Text string `json:"text"`
			From string `json:"from"`
		}
	
		unmarshalErr := json.Unmarshal([]byte(chatHistory), &messages)
		if unmarshalErr != nil {
			var singleMessage struct {
				Text string `json:"text"`
				From string `json:"from"`
			}
			if err2 := json.Unmarshal([]byte(chatHistory), &singleMessage); err2 == nil {
				messages = append(messages, singleMessage)
			} else {
				return "", nil, fmt.Errorf("erro ao processar histórico: %v", unmarshalErr)
			}
		}
	
		for _, msg := range messages {
			if msg.Text == "" {
				continue
			}
			from := msg.From
			if from == "assistant" {
				from = "assistant"
			}
			formattedHistory = append(formattedHistory, fmt.Sprintf("%s: %s", from, msg.Text))
		}
	}
	fmt.Println("Mensagem formatada:", formattedHistory)	

	var embedding any
	var _err error
	if history == "Y" {
		// 1.0 Caso haja historico cria uma pergunta com base na pergunta atual e no historico do chat
		generateQuestion, err := s.gemini.GenerateQuestionByHist(question, formattedHistory)
		if err != nil {
			return "", nil, fmt.Errorf("erro ao gerar pergunta: %v", err)
		}
		fmt.Println(generateQuestion)

		// Remove crases e "json" caso existam (limpeza)
		clean := strings.TrimSpace(generateQuestion)
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")

		// Extrai o conteúdo do "response"
		var respObj map[string]string
		if err := json.Unmarshal([]byte(clean), &respObj); err != nil {
			return "", nil, fmt.Errorf("erro ao extrair response: %v", err)
		}

		questionToEmbedding := respObj["response"]
		fmt.Println(questionToEmbedding)

		// 1.1 Gera embedding com Gemini
		embedding, _err = s.gemini.GenerateEmbedding(questionToEmbedding)
		if _err != nil {
			return "", nil, fmt.Errorf("erro ao gerar embedding: %v", err)
		}
	} else {
		// 1.1 Gera embedding com Gemini
		embedding, _err = s.gemini.GenerateEmbedding(question)
		if _err != nil {
			return "", nil, fmt.Errorf("erro ao gerar embedding: %v", _err)
		}
	}	
	fmt.Println("Embedding:", embedding)
	// 2. Busca documentos similares
	docs, err := s.supabase.FindSimilarDocuments(embedding.([]float32))
	if err != nil {
		return "", nil, fmt.Errorf("erro ao buscar documentos: %v", err)
	}
	fmt.Printf("documentos encontrados", docs)

	// 3. Prepara contexto
	var contextDocs []string
	for _, doc := range docs {
		var docInfo string
		for key, value := range doc {
			docInfo += fmt.Sprintf("%s: %v | ", key, value)
		}

		if len(docInfo) > 0 {
			docInfo = docInfo[:len(docInfo)-2]
		}

		contextDocs = append(contextDocs, docInfo)
	}

	// 4. Gera SQL com Gemini
	sqlQuery, err := s.gemini.GenerateSQL(question, formattedHistory, contextDocs)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao gerar SQL: %v", err)
	}

	fmt.Printf("Retorno da geração: %v", sqlQuery)
	
	// 5. Executa query
	result, err := s.supabase.ExecuteQuery(sqlQuery)
	if err != nil {
		return sqlQuery, nil, fmt.Errorf("erro ao executar query: %v", err)
	}

	if resultMap, ok := result.(map[string]interface{}); ok {
		if salute, ok := resultMap["salute"].(bool); ok && salute {
			fmt.Println("Salute é true — tomando uma ação diferente")

			if msg, ok := resultMap["message"].(string); ok {
				fmt.Println("Mensagem:", msg)
				data := msg

				return sqlQuery, data, nil
			}
		}
	}

	fmt.Printf("Retorno dos itens: %v", result)

	var sb strings.Builder

	items, ok := result.([]map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("tipo de resultado inesperado: %T", result)
	}

	for _, item := range items {
		sb.WriteString("- ")
		primeira := true
		for k, v := range item {
			if !primeira {
				sb.WriteString(" | ")
			}
			sb.WriteString(fmt.Sprintf("%v: %v", k, v))
			primeira = false
		}
		sb.WriteString("\n")
	}	

	formattedResult := sb.String()
	fmt.Printf("formattedResult: %v", formattedResult)

	// 6. Retorna resultado
	resultIA, err := s.gemini.GenerateResponse(question, formattedResult)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao gerar resposta: %v", err)
	}

	return sqlQuery, resultIA, nil
}
