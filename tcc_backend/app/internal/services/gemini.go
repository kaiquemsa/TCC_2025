package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kaiquemsa/nlp-sql-backend/app/handlers"
	"github.com/kaiquemsa/nlp-sql-backend/app/types"
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
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	context := ""
	context = fmt.Sprintf(`
		<CONTEXTO>
		Você atua como otimizador de perguntas. Sua função é analisar o histórico de conversas e a última pergunta do usuário para criar uma versão otimizada da pergunta, clara, completa e independente do contexto anterior.
		</CONTEXTO>
		
		<REGRAS>
		1. A sessão QUESTION contém a última pergunta do usuário.
		2. A sessão HISTORY contém as últimas interações entre o usuário (me) e o atendente (assistant).
		3. Seu objetivo é interpretar a pergunta atual levando em conta o histórico, mas dar prioridade ao conteúdo de QUESTION.
		4. Se a pergunta em QUESTION for ambígua ou depender de contexto do histórico (ex: "E desses, qual foi o último?"), use o HISTORY para completá-la.
		5. Se o histórico tiver múltiplos assuntos, use apenas o mais recente e relevante.
		6. O resultado final deve ser uma única pergunta clara e objetiva, que possa ser usada diretamente para consulta em base de dados ou sistemas.
		7. Nunca responda à pergunta. Seu papel é apenas transformá-la em uma versão otimizada da questão central.
		</REGRAS>
		
		<FORMATACAO>
		Responda sempre em JSON no formato:
		{
			"analise": "(explique passo a passo como interpretou o histórico e a pergunta)",
			"response": "(aqui a pergunta otimizada, em português claro, completa e independente)"
		}
		</FORMATACAO>
		
		<QUESTION>
		"%s"
		</QUESTION>
		
		<HISTORY>
		"%s"
		</HISTORY>
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
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

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
	Lembre-se que o banco de dados é um Postgre, então o SQL criado tem que ser compativel com esse banco.
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
		"salute": (aqui o boolean se true ou false a depender da sua analise se é uma saudação, nunca string),
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("erro na requisição, código %d, detalhe: %s", resp.StatusCode, string(bodyBytes))
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
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	context := ""
	context = fmt.Sprintf(`
	Voce é um analista de banco de dados, muito animado e tem satisfação em atender e responder com êxito o seu cliente, sua função é avaliar os dados, voce deve pegar os dados da sessão CONTENT e retornar apresentando os dados para o cliente.
	Apresente o resultado da pesquisa SQL contendo os campos de retorno formatados usando HTML e como se tivesse apresentando um relatório, mas seja breve, entregue o resultado e uma pequena introdução do que está se tratando.
	O header da tabela montada deve ser de cor escura e as letras claras.
	Titulos e conteúdos fora da tabela devem ter harmonia em questão de tamanho (h1, h2, h3, h4...), espaçamentos e formatação no geral.
	No final da resposta apenas diga que está disponivel para qualquer duvida, não "assine" nada.
	Caso a pergunta do usuario na sessao PERGUNTA seja uma solicitação de gráfico, tabela, comparação, etc, gere isso a ele por meio do HTML.
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

func (s *GeminiService) GenerateEnvelope(question string, result string) (types.Envelope, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=" + os.Getenv("GEMINI_API_KEY")

	prompt := fmt.Sprintf(`
	Você é um assistente de dados. Responda **APENAS** com um JSON válido (sem Markdown, sem crases), no formato:

	{ "type": "text" | "html" | "chart", 
	"text": "<quando type=text>", 
	"html": "<quando type=html>", 
	"spec": { 
		"kind": "bar|line|pie", 
		"title": "...", 
		"x": ["..."], 
		"series": [ { "name":"...", "values":[<número ou null>] } ], 
		"yLabel": "...", 
		"stacked": true|false, 
		"colors": ["#RRGGBB", ... opcional]
	} 
	}

	Regras:
	- Se **um gráfico** for a melhor saída, retorne { "type":"chart", "spec":{...} }.
	- Se for apenas texto simples, use { "type":"text", "text":"..."}.
	- Se precisar formatação rica (lista, tabela), use { "type":"html", "html":"..."} com HTML básico (sem <script>).
	- NÃO retorne HTML quando "type":"chart".
	- Use dados de CONTENT para popular x/series do gráfico quando for o caso.

	<PERGUNTA>
	%s
	</PERGUNTA>

	<CONTENT>
	%s
	</CONTENT>
	`, question, result)

	req := GeminiRequest{
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
					{Text: prompt},
				},
			},
		},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return types.Envelope{}, fmt.Errorf("erro ao converter dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return types.Envelope{}, fmt.Errorf("erro ao enviar requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return types.Envelope{}, fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Envelope{}, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	var gem GeminiResponse
	if err := json.Unmarshal(body, &gem); err != nil {
		return types.Envelope{}, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}
	if len(gem.Candidates) == 0 || len(gem.Candidates[0].Content.Parts) == 0 {
		return types.Envelope{}, fmt.Errorf("resposta vazia do modelo")
	}

	raw := strings.TrimSpace(gem.Candidates[0].Content.Parts[0].Text)
	// Remove fences se vierem por engano
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	// Tenta decodificar no envelope
	var env types.Envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		// fallback: manda como texto
		return types.Envelope{Type: "text", Text: raw}, nil
	}
	// sanity mínimo
	switch env.Type {
	case "text", "html", "chart":
		// ok
	default:
		env.Type = "text"
		if env.Text == "" && env.Html != "" {
			env.Text = env.Html
			env.Html = ""
		}
	}

	return env, nil
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
