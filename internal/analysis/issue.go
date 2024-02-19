package analysis

// Issue problem in analysis.
type Issue struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Hash     string `json:"hash"`
}
