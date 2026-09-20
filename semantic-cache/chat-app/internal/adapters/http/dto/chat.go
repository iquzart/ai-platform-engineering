package dto

type ChatRequest struct {
	Query string `json:"query"`
}
type ChatResponse struct {
	Answer     string  `json:"answer"`
	CacheHit   bool    `json:"cache_hit"`
	Similarity float64 `json:"similarity"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}
