package transaction

import "context"

// このcontextを使ってRepositoryを呼ぶと自動的に同一transaction内で実行される
// fnが nilを返したら commit, errorを返したら rollback
// postgresのdeadlock検出時は最大３回までリトライする
// すでにtxに参加している場合はそのtxを使う
type ITransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
