package edgerouter

import (
	"fmt"
)

type Component struct {
	ER     EdgeRouter
	Server Server
	Trans  Transport
	Ctrl   Controller
}

func (c Component) String() string {
	return fmt.Sprintf("component controlled by(%s) on (%s)", c.Ctrl, c.Trans)
}
