# Phase 6: Outboxパターン（イベント発行の信頼性担保）

## 目的

DBの状態変更とイベント発行の整合性を保証する。
現状のInMemoryPublisherは「DB保存後、Publish前にサーバー落ちるとイベント消える」リスクがある。
Outboxテーブルを介することで「DB変更が成功すればイベントは必ず発行される」を担保する。

実装方針: **APIサーバー内のgoroutineとしてOutboxワーカーを動かす（コンテナ分離なし）**

---

## 現状の問題

```
現状（InMemoryPublisher）:
  txManager.Run(ctx, func(ctx) error {
    checkInRepo.Save()
    eventPublisher.Publish(CheckInCompleted)  ← in-memory配信
  })
  
  → Subscriber（ポイント付与）はトランザクション内で実行される
  → でも非同期SubscribeAsyncで通知ハンドラはトランザクション外
  → サーバー落ちると非同期側のイベントは消える
```

問題: **DB保存と非同期イベント発行の原子性がない**

---

## Outboxパターンで解決

```
Phase 6（Outbox経由）:
  txManager.Run(ctx, func(ctx) error {
    checkInRepo.Save()
    outboxRepo.Save(event)  ← 同じトランザクション内でDBに保存
  })
  → COMMITした時点でイベント発行は予約済み
  
  別goroutineのOutboxWorker:
    while (true) {
      sleep(5秒)
      未処理のoutbox_eventsを取得（LIMIT 20、FOR UPDATE）
      各イベントをsubscribers（既存のEventPublisher）に配信
      成功 → status='SENT'に更新
      失敗 → retry_count増やす、後でリトライ
    }
```

---

## アーキテクチャ

### ディレクトリ構成

```
domain/
  event/
    outbox.go                # IOutboxRepository インターフェース

infrastructure/
  database/
    outbox_repository.go     # PostgreSQL実装
  event/
    outbox_publisher.go      # IEventPublisher実装、内部でoutboxRepoに保存
    outbox_worker.go         # ポーリングワーカー

migrations/
  XXXXXX_add_outbox_events.sql
```

### 依存関係

```
Application層
  └→ EventPublisher (interface)
        └→ OutboxPublisher (実装)  ← outboxRepoにINSERTするだけ
              └→ IOutboxRepository

OutboxWorker（goroutine）
  └→ IOutboxRepository（取り出し）
  └→ InMemoryPublisher（実際の配信）
```

---

## DBスキーマ

```sql
CREATE TABLE outbox_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type varchar(100) NOT NULL,           -- 'CheckInCompleted', 'CheckInSucceeded'等
  payload jsonb NOT NULL,                     -- イベントデータ（JSONシリアライズ）
  status varchar(20) NOT NULL DEFAULT 'PENDING',  -- 'PENDING', 'SENT', 'FAILED'
  retry_count int NOT NULL DEFAULT 0,
  max_retries int NOT NULL DEFAULT 5,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz,                   -- 送信完了時刻

  -- ポーリング高速化のため
  CONSTRAINT outbox_events_status_created_at_idx
);

CREATE INDEX idx_outbox_events_status_created_at
  ON outbox_events (status, created_at)
  WHERE status = 'PENDING';
```

---

## 実装詳細

### 1. ドメイン層: IOutboxRepository

```
domain/event/outbox.go

type OutboxEntry struct {
    id          OutboxEventID
    eventType   string
    payload     []byte    // JSONシリアライズ済み
    status      OutboxStatus
    retryCount  int
    maxRetries  int
    lastError   *string
    createdAt   time.Time
    processedAt *time.Time
}

type IOutboxRepository interface {
    Save(ctx context.Context, entry OutboxEntry) error
    FindPending(ctx context.Context, limit int) ([]OutboxEntry, error)
    MarkSent(ctx context.Context, id OutboxEventID) error
    MarkFailed(ctx context.Context, id OutboxEventID, errMsg string) error
}
```

### 2. インフラ層: OutboxPublisher

既存のEventPublisherを置き換える。Publishの代わりにoutbox_eventsにINSERTするだけ。

```
infrastructure/event/outbox_publisher.go

type OutboxPublisher struct {
    outboxRepo IOutboxRepository
}

func (p *OutboxPublisher) Publish(ctx context.Context, event DomainEvent) error {
    payload, _ := json.Marshal(event)
    entry := OutboxEntry{
        id:        NewOutboxEventID(),
        eventType: event.EventType(),
        payload:   payload,
        status:    OutboxStatusPending,
    }
    return p.outboxRepo.Save(ctx, entry)
    // ↑ context経由でtxに乗るので、トランザクション内で確実に保存される
}
```

### 3. Outboxワーカー（goroutine）

```
infrastructure/event/outbox_worker.go

type OutboxWorker struct {
    outboxRepo IOutboxRepository
    publisher  IEventPublisher  // 実際の配信先（InMemoryPublisher等）
    interval   time.Duration    // ポーリング間隔（5秒）
    batchSize  int              // 1回の取得数（20件）
}

func (w *OutboxWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return  // graceful shutdown
        case <-ticker.C:
            w.processPendingEvents(ctx)
        }
    }
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
    entries, err := w.outboxRepo.FindPending(ctx, w.batchSize)
    if err != nil {
        log.Error("failed to find pending events", err)
        return
    }

    for _, entry := range entries {
        event := deserializeEvent(entry.eventType, entry.payload)
        if err := w.publisher.Publish(ctx, event); err != nil {
            w.outboxRepo.MarkFailed(ctx, entry.id, err.Error())
            continue
        }
        w.outboxRepo.MarkSent(ctx, entry.id)
    }
}
```

### 4. main.goでの組み立て

```
// 既存
inMemoryPublisher := event.NewInMemoryPublisher()
eventPublisher := event.NewOutboxPublisher(outboxRepo)  // ← 差し替え

// EventハンドラはinMemoryPublisherにSubscribe
inMemoryPublisher.Subscribe(types.EventTypeCheckInCompleted, earnPointsHandler.Handle)

// Outboxワーカーをgoroutineで起動
worker := event.NewOutboxWorker(outboxRepo, inMemoryPublisher, 5*time.Second, 20)
ctx, cancel := context.WithCancel(context.Background())
go worker.Run(ctx)

// graceful shutdown時にcancel()を呼ぶ
defer cancel()
```

---

## 既存実装からの変更点

| ファイル | 変更内容 |
|---------|---------|
| domain/event/outbox.go | 新規作成（インターフェース定義） |
| infrastructure/database/outbox_repository.go | 新規作成（PostgreSQL実装） |
| infrastructure/event/outbox_publisher.go | 新規作成（IEventPublisher実装、Outbox保存のみ） |
| infrastructure/event/outbox_worker.go | 新規作成（ポーリングワーカー） |
| infrastructure/event/in_memory_publisher.go | 既存維持（Workerが内部で使用） |
| migrations/ | outbox_events テーブル追加 |
| cmd/server/main.go | Workerをgoroutineで起動 |
| application/check_in/check_in.go | 変更なし（EventPublisherインターフェース経由） |

依存逆転（Clean Architecture）を保っているので、ApplicationレイヤーやEventハンドラは無変更。

---

## 失敗時の挙動

```
イベント発行失敗:
  retry_count++
  last_error = エラーメッセージ
  status = 'PENDING' のまま（次回ポーリングで再試行）

retry_count >= max_retries:
  status = 'FAILED' に変更
  → ワーカーは拾わなくなる
  → 手動対応 or DLQ的な扱い
```

---

## トランザクション内での重要な動作

```
txManager.Run(ctx, fn) {
  BEGIN
  fn(ctx) {
    checkInRepo.Save(ctx, ...)         // contextからtxを取り出して使う
    outboxPublisher.Publish(event)     // 内部でoutboxRepo.Save(ctx, ...)
                                       // 同じtxを使うので原子的
  }
  COMMIT or ROLLBACK
}
```

**ポイント**: contextからtxを取り出す既存のRepositoryパターンに乗っているので、Outbox INSERTもトランザクション内で実行される。

---

## 実装順序

| ステップ | やること |
|---------|---------|
| 1 | outbox_events テーブルのマイグレーション作成 |
| 2 | OutboxEntry エンティティ + Value Objects |
| 3 | IOutboxRepository インターフェース定義 |
| 4 | PostgreSQL実装（Save, FindPending, MarkSent, MarkFailed） |
| 5 | OutboxPublisher（IEventPublisher実装、INSERTのみ） |
| 6 | OutboxWorker（ポーリングロジック） |
| 7 | main.goでgoroutine起動 + EventPublisher差し替え |
| 8 | テスト（ハッピーパス、リトライ、サーバー停止時の永続化確認） |

---

## 面接での語り方

> 「StreekでOutboxパターンを実装しました。EventPublisherを直接呼ぶと、DB保存とイベント発行の原子性が崩れる（DB保存後・発行前にサーバー落ちるとイベント消える）問題があります。
>
> Outboxテーブルにイベントを「予約」として同じトランザクション内で保存し、別goroutineのOutboxWorkerが5秒ごとにポーリングして発行する設計にしました。これでDB変更が成功すればイベントは必ず発行されることが保証されます。
>
> KalonadeのMessenger ServiceやMercoinのRecovery Workerと同じ思想で、金融システムでイベント発行の信頼性を担保する標準パターンです」

---

## 注意点

- ポーリング間隔（5秒）は遅延とDB負荷のトレードオフ。本番なら1秒程度が普通
- バッチサイズ（20件）も同様。多すぎるとロック時間が長くなる
- 失敗時の指数バックオフは将来追加（今回は単純リトライ）
- イベントの順序保証は不要（CheckInCompletedとCheckInSucceededは独立）
- 重複発行に備えて、Subscriber側で冪等性を担保する（既に実装済み: ポイント付与の冪等性キー）
