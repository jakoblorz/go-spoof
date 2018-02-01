package tunnel

import "fmt"

// Endpoint represents a single Server
// with Hostname and Port
type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String(network string) string {
	return fmt.Sprintf("%s://[%s]:%d", network, endpoint.Host, endpoint.Port)
}
