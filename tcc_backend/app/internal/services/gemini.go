package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/kaiquemsa/nlp-sql-backend/app/handlers"
)

type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
		Role string `json:"role"`
	} `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func GenerateQuestionByHist(question string, history []string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	context := ""
	context = fmt.Sprintf(`
	Você é um analista de dados responsável por transformar conversas em perguntas claras para busca em banco de dados.  
	Receba:
	
	<QUESTION>
	"%s"
	</QUESTION>
	
	<HISTORY>
	"%s"
	</HISTORY>
	
	Com base no histórico (interação entre "me" e "assistant") e na pergunta atual, **elabore uma nova pergunta clara, completa e desambígua**, capaz de ser usada diretamente para buscar informações relevantes no banco de dados (ex: ao invés de "repita a resposta anterior", escreva a pergunta anterior explicitamente).  
	**Não explique nada, apenas retorne o resultado no formato abaixo:**  
	
	<EXAMPLE>
	{
	  "response": "Aqui vai a pergunta gerada, pronta para gerar um embedding e buscar no banco"
	}
	</EXAMPLE>
	`, question, history)	

	requestData := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		}{
			{
				Role: "user",
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: context},
				},
			},
		},
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição para Gemini: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na requisição, código de status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler a resposta: %v", err)
	}

	var gemResp GeminiResponse
	if err := json.Unmarshal(bodyBytes, &gemResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	if len(gemResp.Candidates) == 0 {
		return "", fmt.Errorf("nenhuma resposta do Gemini")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}

func GenerateSQL(question string, history []string, contextDocs []string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	context := ""
	context = fmt.Sprintf(`
	Voce é um analista de banco de dados, sua função é avaliar os dados e suas similaridades na sessão CONTENT, voce deve avaliar os dados e usar os mais similares possiveis comparados a pergunta do usuário na sessão PERGUNTA e montar um SQL de acordo com os conteudos que estejam de acordo com o que foi perguntado e no que contem nos documentos apresentados. 
	Com base nisso, escreva uma única consulta SQL simples que pode ser usada para buscar esse conteúdo objetivo da PERGUNTA no banco de dados. Não inclua explicações, apenas o SQL.
	Use a sessão TABLE para não cometer erros na montagem do SQL.
	Caso voce identifique que se trata apenas de uma saudação ou uma mensagem que não seja uma pergunta/pedido, adicione um retorne de uma flag "salute": "true". 
	Caso haja conteúdo dentro da sessão HISTORY analise se há relevancia em relação a pergunta antes de formar a resposta.
	Use apenas o conteúdo disponivel dentro da sessão CONTENT e TABLE, não invente nome de tabelas e nem de colunas.
	Coloque espaço entre as palavras da query SQL montado, não deixe os comandos colados, pois pode dar problemas.
	Coloque a tabela e a(s) coluna(s) entre aspas sempre.
	Tenha certeza de que a tabela e colunas indicadas no SQL montado estejam dentro da sessão CONTENT.
	Não crie funções.
	Não coloque ";" no final da query.
	Formate a saída da resposta de acordo com a sessão EXEMPLO.
	Revise o SQL 3x antes de retornar o resultado, não coloque tabelas e colunas que não estão dentro da sessão CONTENT.
	<PERGUNTA>
	Pergunta do usuário: "%s".
	</PERGUNTA>
	<HISTORY>
	Historico da conversa a ser considerado para análise: "%s",
	</HISTORY>
	<CONTENT>
	Considere o seguinte conteúdo:
	"%s".
	</CONTENT>
	<TABLE>
	Estrutura de todas as tabelas e suas colunas:
	ordem_producao: [id: int8, numero_ordem_de_producao: text, data_de_criacao: timestamptz, sku_id: int8, ordem_ativa: text, numero_de_pallets_saida: int8, numero_de_pallets_produzidos: int8, quantidade: int8]
	pallets_saida: [id: int8, data_de_criacao: timestamptz, quantidade: int8, ordem_de_producao_id: int8]
	linha_producao: [id: int8, nome_da_linha: text, cd_linha: text, cd_status: int8]
	sku: [id: int8, sku_number: text, codigo_pais: text, descricao: text, quantidade_de_itens_pallet_saida: int8, cd_status: int8]
	grupo_usuarios: [id: int8, nome_do_grupo: text, cd_status: int8]
	usuarios: [id: int8, username: text, nome: text, conta_habilitada: bytea, grupo_de_usuario_id: int8]
	</TABLE>
	<EXEMPLO>
	{
		"sql": (aqui o resultado do sql),
		"salute": (aqui se true ou false a depender da sua analise se é uma saudação),
		"response": (aqui caso salute for true, escreva uma resposta de saudação de acordo com a mensagem do usuário e ofereça sua ajuda)
	}
	</EXEMPLO>
	`, question, history, contextDocs)

	requestData := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		}{
			{
				Role: "user",
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: context},
				},
			},
		},
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição para Gemini: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na requisição, código de status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler a resposta: %v", err)
	}

	var gemResp GeminiResponse
	if err := json.Unmarshal(bodyBytes, &gemResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	if len(gemResp.Candidates) == 0 {
		return "", fmt.Errorf("nenhuma resposta do Gemini")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}

func GenerateResponse(question string, result string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	context := ""
	context = fmt.Sprintf(`
	Voce é um analista de banco de dados, muito animado e tem satisfação em atender e responder com êxito o seu cliente, sua função é avaliar os dados, voce deve pegar os dados da sessão CONTENT e retornar apresentando os dados para o cliente.
	Apresente o resultado da pesquisa SQL contendo os campos de retorno formatados usando HTML e como se tivesse apresentando um relatório, mas seja breve, entregue o resultado e uma pequena introdução do que está se tratando.
	O header da tabela montada deve ser de cor escura e as letras claras.
	Titulos e conteúdos fora da tabela devem ter harmonia em questão de tamanho (h1, h2, h3, h4...), espaçamentos e formatação no geral.
	No final da resposta apenas diga que está disponivel para qualquer duvida, não "assine" nada.
	<PERGUNTA>
	Pergunta do usuário: "%s".
	</PERGUNTA>
	<CONTENT>
	Considere o seguinte conteúdo:
	"%s".
	</CONTENT>
	`, question, result)

	requestData := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		}{
			{
				Role: "user",
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: context},
				},
			},
		},
	}

	payload, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("erro ao enviar requisição para Gemini: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erro na requisição, código de status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler a resposta: %v", err)
	}

	var gemResp GeminiResponse
	if err := json.Unmarshal(bodyBytes, &gemResp); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	if len(gemResp.Candidates) == 0 {
		return "", fmt.Errorf("nenhuma resposta do Gemini")
	}

	return gemResp.Candidates[0].Content.Parts[0].Text, nil
}

// Serviço Gemini
type GeminiService struct{}

func NewGeminiService() *GeminiService {
	return &GeminiService{}
}

func (g *GeminiService) GenerateEmbedding(question string) (any, error) {
	fmt.Printf("Pergunta recebida: ", question)
	embedding, err := handlers.GenerateEmbeddingLocal(question)

	if err != nil {
		return nil, err
	}

	return embedding, nil
}


// Função que usa a API Gemini para gerar o SQL
func (g *GeminiService) GenerateSQL(question string, history []string, contextDocs []string) (string, error) {
	return GenerateSQL(question, history, contextDocs)
}

func (g *GeminiService) GenerateResponse(question string, result string) (string, error) {
	return GenerateResponse(question, result)
}

func (g *GeminiService) GenerateQuestionByHist(question string, history []string) (string, error) {
	return GenerateQuestionByHist(question, history)
}
