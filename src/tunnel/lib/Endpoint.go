package tunnel

import "fmt"

// Endpoint represents a single Server
// with Hostname and Port
type Endpoint struct {
	Host string
	Port int
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}
