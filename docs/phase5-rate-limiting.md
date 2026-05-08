# Phase 5: Rate Limiting（スライディングウィンドウ / Redis Sorted Set）

## 目的

ユーザーごとにAPIの呼び出し回数を制限し、不正利用やシステム過負荷を防止する。
金融システムではアプリ層でのレートリミットが必須。

## アルゴリズム: スライディングウィンドウ

固定ウィンドウの境界問題（ウィンドウの端でまとめてアクセスすると制限を超えられる）を防ぐため、スライディングウィンドウを採用。「直近N秒間」で常に評価するので抜け道がない。

---

## アーキテクチャ

### ディレクトリ構成

```
domain/
  ratelimit/
    rate_limiter.go              # IRateLimiter インターフェース

infrastructure/
  ratelimit/
    redis_rate_limiter.go        # Redis Sorted Set 実装
    redis_rate_limiter_test.go   # テスト

middleware/
    rate_limit.go                # Echo ミドルウェア
```

### 依存方向

```
middleware/rate_limit.go → domain/ratelimit/IRateLimiter ← infrastructure/ratelimit/RedisRateLimiter
```

middlewareはドメインのインターフェースに依存。Redis実装はインフラ層。依存逆転を維持。

### リクエストの流れ

```
リクエスト
  → auth middleware（JWT検証、UserID取得）
  → rate limit middleware（ユーザーIDでレート判定）
     → 制限内 → handler → service → repository → レスポンス
     → 制限超過 → 429 Too Many Requests を即返す
```

rate limit middleware は auth middleware の後に配置する。UserIDが必要なため。

---

## インターフェース

### IRateLimiter（domain/ratelimit/rate_limiter.go）

```go
type IRateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
}
```

- key: 制限の単位（ユーザーIDを渡す）
- 戻り値: true = 許可、false = 制限超過
- エラー: Redis障害時など。エラー時はリクエストを通す（後述）

---

## Redis の基礎知識（前提）

### Redis とは
- メモリ上でデータを管理する超高速なデータベース（Key-Valueストア）
- PostgreSQLがディスクに保存するのに対し、Redisはメモリに保存 → 読み書きが非常に速い
- Streekでは既にストリークキャッシュで使用済み。同じ接続を再利用する

### Sorted Set（ソーテッドセット）とは
- Redisに最初から入っているデータ構造の1つ
- 「値」と「スコア（点数）」のペアを保存する
- スコアの順に自動でソートされる
- 自分で作るものではなく、Redisのコマンドで操作するだけ

```
普通の配列:
  ["りんご", "みかん", "バナナ"]
  → 順番はあるが、点数はない

Sorted Set:
  { "りんご": 100, "みかん": 200, "バナナ": 150 }
  → スコア順に自動ソート: りんご(100) → バナナ(150) → みかん(200)
```

### 今回の使い方
- スコア = リクエストの時刻（UnixNano）
- 値 = リクエストの時刻（文字列。一意にするため）
- 時刻順にソートされるので「何秒前より古いデータを削除」が1コマンドでできる

```
Sorted Set "ratelimit:user123":

  値（メンバー）          スコア（時刻）
  "1715100000000"     1715100000000   ← 1回目のリクエスト
  "1715100001000"     1715100001000   ← 2回目のリクエスト
  "1715100003000"     1715100003000   ← 3回目のリクエスト
```

### 使用する Redis コマンド（4つだけ）

| コマンド | やること | 例 |
|---------|---------|---|
| ZADD | Sorted Set に要素を追加 | 「user123のセットにタイムスタンプを追加」 |
| ZREMRANGEBYSCORE | スコア範囲を指定して削除 | 「60秒より前のデータを全部消す」 |
| ZCARD | Sorted Set の要素数を取得 | 「今セットに何件ある？」→ 8件 |
| EXPIRE | キーに有効期限を設定 | 「このキーを60秒後に自動削除」 |

### Redis Pipeline とは
- 通常: コマンドを1つ送る → 結果を待つ → 次のコマンドを送る（4往復）
- Pipeline: コマンドを4つまとめて送る → 結果をまとめて受け取る（1往復）
- ネットワークの往復が減るので速い。今回は4コマンドを1回のPipelineで実行する

---

## Redis実装

### RedisRateLimiter（infrastructure/ratelimit/redis_rate_limiter.go）

コンストラクタの引数:

| パラメータ | 型 | 説明 | Streekでの値 |
|----------|---|------|------------|
| client | *redis.Client | 既存のRedisクライアントを再利用 | main.goで生成済み |
| limit | int | ウィンドウ内の最大リクエスト数 | 10 |
| window | time.Duration | ウィンドウの長さ | 1 * time.Minute |
| now | func() time.Time | 時刻関数（テスト用に注入可能） | time.Now |

### Allow() の処理（4ステップ）

Redis Sorted Set を使い、スコア = UNIXタイムスタンプ（ナノ秒）で管理する。

```
キー: "ratelimit:{userID}"
値:   Sorted Set（スコア = リクエスト時刻のUnixNano）

Step 1: ZREMRANGEBYSCORE で古いデータを削除
  → now - window より前のエントリを全削除
  → これにより「直近N秒」のデータだけが残る

Step 2: ZCARD で現在のリクエスト数をカウント
  → Sorted Set の要素数 = 直近N秒間のリクエスト数

Step 3: 制限内か判定
  → count >= limit なら false を返す（制限超過）
  → count < limit なら Step 4 へ

Step 4: ZADD で今のリクエストを追加
  → スコア = now.UnixNano()
  → メンバー = now.UnixNano() の文字列（一意にするため）
  → キーに EXPIRE を設定（window の長さ。メモリリーク防止）
```

### Redisコマンドの実行順

4つのコマンドを1回のリクエストで実行する必要がある。Redis Pipeline を使ってラウンドトリップを1回にする。

```
pipe := client.Pipeline()
pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
countCmd := pipe.ZCard(ctx, key)
pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
pipe.Expire(ctx, key, window)
_, err := pipe.Exec(ctx)
```

ただし、ZADDをcountの判定前に実行してしまうとcount+1で判定することになるので注意。
方法は2つ:
- A) Pipeline を2回に分ける（ZREMRANGEBYSCORE + ZCARD → 判定 → ZADD + EXPIRE）
- B) Pipeline 1回で全部実行し、count の判定を `count >= limit` ではなく `count > limit` にする（ZADDが先に入るのでcountが1多い）

Bの方がシンプルでラウンドトリップも1回なのでおすすめ。

Bの場合の流れ:
```
pipe.ZRemRangeByScore(...)   // 古いの削除
pipe.ZAdd(...)               // 先に追加
countCmd = pipe.ZCard(...)   // 追加後のカウント
pipe.Expire(...)             // TTL設定
pipe.Exec(ctx)

count := countCmd.Val()
return count <= int64(limit)  // 追加後なので <= で判定
```

### エラー時の方針: フェイルオープン

Redis障害時はリクエストを**通す**（429にしない）。

理由: Redisが落ちたせいで正規ユーザーが全員ブロックされるのは本末転倒。レートリミットはあくまで保護機能であり、本体のサービスを止めてはいけない。

```go
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
    // ... Redis操作 ...
    if err != nil {
        // Redis障害 → 通す（ログは出す）
        return true, err
    }
    return count <= int64(r.limit), nil
}
```

middleware側でもerrorを受け取ったらログを出して通す。

---

## ミドルウェア

### RateLimitMiddleware（middleware/rate_limit.go）

```
func RateLimitMiddleware(limiter ratelimit.IRateLimiter) echo.MiddlewareFunc
```

処理:
1. contextからUserIDを取得（auth middleware が設定済み）
2. limiter.Allow(ctx, userID.String()) を呼ぶ
3. false → 429 Too Many Requests をレスポンス
4. true → next(c) で次のハンドラへ
5. error → ログ出力して next(c)（フェイルオープン）

429レスポンスの形式:
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "too many requests, please try again later"
  }
}
```

apperror にエラーコードを追加する:
```go
CodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
```

---

## main.go での組み立て

```
// 既存の redisClient を再利用
rateLimiter := ratelimit.NewRedisRateLimiter(redisClient, 10, 1*time.Minute)

// auth middleware の後に rate limit middleware を適用
habits := api.Group("/habits",
    middleware.AuthMiddleware(tokenGenerator),
    middleware.RateLimitMiddleware(rateLimiter),
)
```

認証不要のエンドポイント（/auth/register, /auth/login）にはレートリミットを適用しない（UserIDがないため）。
将来的にIPベースの制限を追加することは可能だが、今回のスコープ外。

---

## 設定値

| 設定 | 値 | 理由 |
|------|---|------|
| limit | 10 | 1分間に10回。習慣アプリとして妥当 |
| window | 1分 | 短すぎず長すぎない |
| Redisキー | "ratelimit:{userID}" | ユーザーごとに制限 |
| TTL | window と同じ（1分） | 最後のリクエストからwindow経過後にキーが自動削除 |

---

## テスト

### redis_rate_limiter_test.go

テスト用にnowを注入できるので、time.Sleepなしで検証可能。

テストケース:
1. 制限内のリクエスト → Allow = true
2. 制限超過（11回目）→ Allow = false
3. ウィンドウ経過後 → カウントリセット、Allow = true に戻る
4. 異なるユーザーは独立してカウント

テスト環境: docker-compose のRedisを使う（実際のRedisに対してテスト）。

### middleware のテスト（オプション）

IRateLimiter のモックを使って:
1. Allow = true → 次のハンドラが呼ばれる
2. Allow = false → 429 が返る
3. Allow でエラー → 次のハンドラが呼ばれる（フェイルオープン）

---

## 実装順序

| ステップ | やること |
|---------|---------|
| 1 | domain/ratelimit/rate_limiter.go — IRateLimiter インターフェース定義 |
| 2 | apperror/error_code.go — CodeRateLimitExceeded 追加 |
| 3 | infrastructure/ratelimit/redis_rate_limiter.go — Redis Sorted Set 実装 |
| 4 | infrastructure/ratelimit/redis_rate_limiter_test.go — テスト |
| 5 | middleware/rate_limit.go — Echo ミドルウェア |
| 6 | main.go — 組み立て（redisClient再利用、middleware適用） |
| 7 | 手動テスト — curl で11回叩いて429が返ることを確認 |

---

## 面接での語り方

> 「ユーザーごとのAPIレートリミットをRedis Sorted Setのスライディングウィンドウで実装しました。固定ウィンドウの境界問題を避けるためにスライディングウィンドウを選び、Redis Pipelineでラウンドトリップを1回に抑えています。Redis障害時はフェイルオープンでリクエストを通す方針にしました。レートリミットの障害で正規ユーザーをブロックするのは本末転倒だからです」
