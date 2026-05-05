package transaction

import "context"

// ITransactionManager は複数の Repository 操作を一つの DB トランザクションに束ねる。
//
// fn 内で渡される ctx を Repository に渡すと、自動的に同一トランザクション内で実行される。
// fn が nil を返したら commit、error を返したら rollback。
// PostgreSQL の deadlock (SQLSTATE 40P01) 検出時は最大 3 回まで自動リトライ。
// 既に tx に参加している ctx で呼ばれた場合は既存 tx を再利用する（新規 tx は開始しない）。
type ITransactionManager interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}
