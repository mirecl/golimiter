package analysis

// Issue problem in analysis.
type Issue struct {
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Hash     string `json:"hash"`
	Severity string `json:"severity"`
	Type     string `json:"type"`
}
