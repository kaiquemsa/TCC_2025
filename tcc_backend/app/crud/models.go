package crud

type Embedding struct {
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

type Chat struct {
	From    string `json:"from"`
	Text    string `json:"text"`
	Id_chat string `json:"id_chat"`
}
