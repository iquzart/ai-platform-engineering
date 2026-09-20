package usecases

import "time"

type PingResponse struct {
	Service   string
	Timestamp time.Time
}

type Ping struct{ serviceName string }

func NewPing(serviceName string) Ping { return Ping{serviceName: serviceName} }

func (p Ping) Execute() PingResponse {
	return PingResponse{Service: p.serviceName, Timestamp: time.Now().UTC()}
}
