package edgerouter

import "context"

type Server interface {
	Run(ctx context.Context, handler interface{}) (context.Context, error)
}
