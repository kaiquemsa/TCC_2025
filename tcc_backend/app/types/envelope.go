package types

// Tipos suportados pelo frontend
type ChartKind string

const (
	ChartBar  ChartKind = "bar"
	ChartLine ChartKind = "line"
	ChartPie  ChartKind = "pie"
)

type ChartSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

type ChartSpec struct {
	Kind    ChartKind     `json:"kind"`
	Title   string        `json:"title,omitempty"`
	X       []string      `json:"x"`
	Series  []ChartSeries `json:"series"`
	YLabel  string        `json:"yLabel,omitempty"`
	Stacked bool          `json:"stacked,omitempty"`
	Colors  []string      `json:"colors,omitempty"`
}

// Envelope unificado que o front entende
type Envelope struct {
	Type string     `json:"type"`           // "text" | "html" | "chart"
	Text string     `json:"text,omitempty"` // quando type=="text"
	Html string     `json:"html,omitempty"` // quando type=="html"
	Spec *ChartSpec `json:"spec,omitempty"` // quando type=="chart"
}
