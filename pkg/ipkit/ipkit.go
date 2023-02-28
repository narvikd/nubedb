package ipkit

import "fmt"

func NewAddr(nodeID string, port int) string {
	return fmt.Sprintf("%s:%v", nodeID, port)
}
