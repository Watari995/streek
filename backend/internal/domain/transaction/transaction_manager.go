package transaction

import "context"

type ITransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
