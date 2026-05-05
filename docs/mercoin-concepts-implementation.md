# Mercoin 概念の Streek への実装計画

Mercoin面接で「金融システムの基本概念をGoで実装した経験がある」と語るための実装ドキュメント。

CLAUDE.mdのAIルールに従い、Goコードは全て手書きで実装する。

---

## 実装する3つの概念

| 順番 | 概念 | 面接で語れること |
|------|------|--------------|
| 1 | トランザクションマネージャー | 複数のDB操作を原子的に管理できる |
| 2 | ポイントシステム（台帳パターン） | 金融システムの基本である台帳ベースの残高管理を実装した |
| 3 | イベント発行パターン | チェックインとポイント付与を疎結合にし、将来のPub/Sub化に備えた設計 |

---

## Phase 1: トランザクションマネージャー

### なぜ必要か

現在のStreekは各Repositoryが個別にDBアクセスしている。
Phase 2でポイントシステムを追加すると「チェックイン + ポイント付与」を原子的に行う必要がある。
片方だけ成功してもう片方が失敗する状態は、金融システムでは許されない。

### 設計

Domain層にインターフェースを定義し、Infrastructure層でsqlxのトランザクションを使って実装する。

```
domain/transaction/
  └── transaction_manager.go   # ITransactionManager インターフェース

infrastructure/database/
  └── transaction_manager.go   # sqlx.Tx を使った実装
```

### インターフェース

```go
// domain/transaction/transaction_manager.go
type ITransactionManager interface {
    Run(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 仕組み

- Run() の中で sqlx.BeginTxx() でトランザクション開始
- context に tx を埋め込む
- fn 内で呼ばれる Repository は context から tx を取り出して使う
- fn が error を返したら Rollback、nil なら Commit
- デッドロック（serialization failure）時は自動リトライ（最大3回）

### Repository の変更

各Repositoryに「context から tx を取り出す。なければ通常のDB接続を使う」ロジックを追加。

```go
func (r *HabitRepository) getConn(ctx context.Context) sqlx.ExtContext {
    if tx, ok := GetTx(ctx); ok {
        return tx
    }
    return r.db
}
```

### 面接での語り方

> 「context に tx を埋め込むことで、Service層がトランザクションの境界を決め、Repository層は自動的にそのトランザクションに参加する設計にしました。依存逆転を崩さずにトランザクション管理を実現しています」

---

## Phase 2: ポイントシステム（台帳パターン）

### なぜ台帳パターンか

残高を直接持つ（balance カラム）のではなく、全取引を履歴として記録し、残高はその合計で算出する。

- 監査性: いつ・いくら・なぜ増減したかが全て追跡できる
- バグ検知: 二重加算や不正な減算を履歴から検出できる
- 金融の標準: 銀行の通帳と同じ発想

### DBスキーマ

```sql
-- マイグレーションファイル
CREATE TABLE "point_ledger" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "habit_id" uuid NULL,
  "type" varchar(10) NOT NULL,         -- 'EARN' or 'SPEND'
  "amount" integer NOT NULL,            -- 正の数で統一。type で加減を判断
  "reason" varchar(100) NOT NULL,       -- 'CHECK_IN', 'STREAK_BONUS', 'REDEEM' 等
  "idempotency_key" varchar(255) NULL,  -- 冪等性キー（同じキーの重複INSERTを防止）
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "point_ledger_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "point_ledger_idempotency_key_key" UNIQUE ("idempotency_key")
);
CREATE INDEX "idx_point_ledger_user_id" ON "point_ledger" ("user_id");
```

### ポイントの残高算出

```sql
SELECT
  COALESCE(SUM(CASE WHEN type = 'EARN' THEN amount ELSE 0 END), 0) -
  COALESCE(SUM(CASE WHEN type = 'SPEND' THEN amount ELSE 0 END), 0)
  AS balance
FROM point_ledger
WHERE user_id = $1;
```

残高は保存しない。常に計算で出す。
パフォーマンスが必要になったら Redis キャッシュを追加する（Phase 2 の範囲外）。

### Domain層

```
domain/entity/
  └── point_entry.go          # PointEntry エンティティ

domain/valueobject/
  ├── point_amount.go          # 正の整数、0以下はエラー
  ├── point_type.go            # EARN / SPEND の列挙
  └── point_reason.go          # CHECK_IN / STREAK_BONUS / REDEEM 等

domain/repository/
  └── point_ledger_repository.go  # IPointLedgerRepository インターフェース
```

### IPointLedgerRepository

```go
type IPointLedgerRepository interface {
    Save(ctx context.Context, entry entity.PointEntry) (*entity.PointEntry, error)
    GetBalance(ctx context.Context, userID valueobject.UserID) (int, error)
    FindByUserID(ctx context.Context, userID valueobject.UserID) ([]*entity.PointEntry, error)
    ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
}
```

### Application層: チェックイン + ポイント付与

```
application/check_in/
  └── check_in.go   # 既存のチェックインサービスを拡張
```

処理フロー:

```
CheckIn.Do(ctx, input)
  └→ transactionManager.Run(ctx, func(ctx) error {
       ├→ checkInRepo.Save(ctx, checkIn)          // チェックイン記録
       ├→ idempotencyKey を生成（"checkin:{habitID}:{date}"）
       ├→ 冪等性チェック: 既に同じキーがあればスキップ
       └→ pointLedgerRepo.Save(ctx, pointEntry)   // ポイント加算（+10）
     })
```

両方成功するか、両方失敗するか。片方だけ成功することはない。

### 冪等性キー

```
フォーマット: "checkin:{habitID}:{checked_date}"
例: "checkin:550e8400-e29b-41d4-a716-446655440000:2026-05-05"
```

同じ日に同じ習慣のチェックインでポイントが二重付与されることを防ぐ。
check_ins テーブルの UNIQUE 制約とは別に、point_ledger 側でも冪等性を担保する。

### 面接での語り方

> 「ポイントの残高を balance カラムで直接管理するのではなく、台帳パターンで全取引を履歴として記録しています。残高は履歴の合計で算出します。これにより、いつ・いくら・なぜポイントが増減したかを完全に追跡でき、二重加算のバグも冪等性キーで防止しています。銀行の通帳と同じ設計思想です」

---

## Phase 3: イベント発行パターン

### なぜ必要か

Phase 2では「チェックイン + ポイント付与」をService内で直接呼んでいる。
これだと、将来「チェックイン → 通知送信」「チェックイン → ストリークボーナス判定」を追加するたびにServiceが肥大化する。

イベント発行パターンで疎結合にする。

### 設計

最初はGoのchannelでインプロセスに実装する。
将来的にRedis Pub/SubやCloud Pub/Subに差し替えられる設計にしておく。

```
domain/event/
  ├── event.go            # DomainEvent インターフェース
  └── publisher.go        # IEventPublisher インターフェース

domain/event/types/
  └── check_in_completed.go  # CheckInCompletedEvent 構造体

infrastructure/event/
  └── in_memory_publisher.go  # channel ベースの実装

application/event_handler/
  ├── earn_points_on_check_in.go   # ポイント付与ハンドラ
  └── notify_streak_milestone.go   # ストリーク達成通知ハンドラ（将来）
```

### 処理フロー

```
CheckIn.Do(ctx, input)
  └→ transactionManager.Run(ctx, func(ctx) error {
       ├→ checkInRepo.Save(ctx, checkIn)
       └→ eventPublisher.Publish(ctx, CheckInCompletedEvent{...})
     })

↓ イベントを購読しているハンドラが反応

EarnPointsOnCheckIn.Handle(ctx, event)
  └→ pointLedgerRepo.Save(ctx, pointEntry)
```

### 面接での語り方

> 「チェックインとポイント付与を直接結合させるのではなく、ドメインイベントを介して疎結合にしました。今はインプロセスのchannelですが、IEventPublisherインターフェースを切っているので、Cloud Pub/Subへの差し替えが可能です。MercoinのPub/Subによるマイクロサービス間のイベント駆動通信と同じ設計思想です」

---

## ディレクトリ追加（最終形）

```
backend/internal/
├── domain/
│   ├── entity/
│   │   ├── point_entry.go          # NEW
│   │   └── ...
│   ├── valueobject/
│   │   ├── point_amount.go         # NEW
│   │   ├── point_type.go           # NEW
│   │   ├── point_reason.go         # NEW
│   │   └── ...
│   ├── repository/
│   │   ├── point_ledger_repository.go  # NEW
│   │   └── ...
│   ├── transaction/
│   │   └── transaction_manager.go  # NEW: ITransactionManager
│   └── event/
│       ├── event.go                # NEW: DomainEvent interface
│       ├── publisher.go            # NEW: IEventPublisher interface
│       └── types/
│           └── check_in_completed.go  # NEW
├── application/
│   ├── check_in/
│   │   └── check_in.go            # MODIFIED: tx + ポイント付与
│   ├── point/
│   │   ├── get_balance.go          # NEW
│   │   └── get_history.go          # NEW
│   └── event_handler/
│       └── earn_points_on_check_in.go  # NEW
└── infrastructure/
    ├── database/
    │   ├── transaction_manager.go  # NEW: sqlx.Tx実装
    │   ├── point_ledger_repository.go  # NEW
    │   └── ...
    └── event/
        └── in_memory_publisher.go  # NEW
```

---

## Phase 4: サーキットブレーカー + メール通知

### なぜ必要か

ストリーク達成時にメール通知を送りたいが、Gmail SMTPが落ちていたり遅延している場合に、チェックインのレスポンスまで巻き込まれてはいけない。

サーキットブレーカーで：
- Gmail SMTPが連続で失敗したらリクエストを遮断（無駄な待ちを避ける）
- 一定時間後に再試行して回復を検知
- 通知の失敗がチェックイン処理に影響しない設計

### 設計

通知はトランザクション外で実行する（通知失敗でチェックインがロールバックされてはいけない）。

```
チェックイン
  └→ トランザクション内:
       ├→ check_in 保存
       └→ イベント発行 → ポイント付与

  └→ トランザクション外（成功後）:
       └→ ストリーク判定 → N日連続達成？
            └→ サーキットブレーカー経由で Gmail SMTP 通知
                 「7日連続達成！」
```

Mercoinでも決済処理（トランザクション内）と通知送信（トランザクション外）は分離されている。

### サーキットブレーカーの3つの状態

```
CLOSED（通常状態）
  → リクエストを通す
  → 失敗が閾値（例: 3回連続）に達したら OPEN へ

OPEN（遮断状態）
  → リクエストを即座にエラーで返す（外部APIを呼ばない）
  → 一定時間（例: 30秒）経過したら HALF_OPEN へ

HALF_OPEN（試行状態）
  → 1リクエストだけ通して様子を見る
  → 成功したら CLOSED に戻る
  → 失敗したら OPEN に戻る
```

### ディレクトリ構成

```
domain/
  notification/
    notifier.go              # INotifier インターフェース

infrastructure/
  notification/
    email_notifier.go        # Gmail SMTP 実装
  circuitbreaker/
    circuit_breaker.go       # サーキットブレーカー実装
    circuit_breaker_test.go  # 状態遷移テスト

application/
  event_handler/
    notify_streak_milestone.go  # ストリーク達成通知ハンドラ
```

### INotifier インターフェース

```go
// domain/notification/notifier.go
type INotifier interface {
    Notify(ctx context.Context, to string, subject string, body string) error
}
```

EmailNotifier が実装し、サーキットブレーカーでラップする。
インターフェースを切っているので、将来Slack/LINE/Pushに差し替え可能。

### サーキットブレーカーの設計

```go
// infrastructure/circuitbreaker/circuit_breaker.go
type CircuitBreaker struct {
    state           State          // CLOSED, OPEN, HALF_OPEN
    failureCount    int            // 連続失敗回数
    failureThreshold int           // OPEN に遷移する閾値（例: 3）
    resetTimeout    time.Duration  // OPEN → HALF_OPEN までの待ち時間（例: 30秒）
    lastFailureAt   time.Time      // 最後に失敗した時刻
    mu              sync.Mutex     // 並行アクセス制御
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    // OPEN なら即エラー（外部APIを呼ばない）
    // CLOSED / HALF_OPEN なら fn() を実行
    // 結果に応じて状態遷移
}
```

### メール通知の実装

```go
// infrastructure/notification/email_notifier.go
// Go標準ライブラリ net/smtp のみ使用
// Gmail SMTP: smtp.gmail.com:587
// アプリパスワード（Googleアカウントの2段階認証で生成）を使用
```

環境変数:
- SMTP_HOST=smtp.gmail.com
- SMTP_PORT=587
- SMTP_USER=your@gmail.com
- SMTP_PASSWORD=アプリパスワード（16文字）
- NOTIFICATION_TO=通知先メールアドレス

### 処理フロー

```
CheckIn.Do()
  └→ txManager.Run() ← トランザクション内
       ├→ checkInRepo.Save()
       └→ eventPublisher.Publish(CheckInCompletedEvent)
            └→ EarnPointsOnCheckIn.Handle() → ポイント付与

  └→ トランザクション成功後:
       └→ eventPublisher.Publish(CheckInSucceededEvent) ← 新イベント
            └→ NotifyStreakMilestone.Handle()
                 ├→ streakService.ComputeCurrentStreak()
                 ├→ 7日達成？
                 └→ circuitBreaker.Execute(func() error {
                      return notifier.Notify(ctx, to, subject, body)
                    })
```

### テスト

サーキットブレーカーのテスト（モックで外部API不要）:
- CLOSED → 3回連続失敗 → OPEN に遷移
- OPEN → リクエストが即エラーで返る（fn が呼ばれない）
- OPEN → 30秒経過 → HALF_OPEN に遷移
- HALF_OPEN → 成功 → CLOSED に戻る
- HALF_OPEN → 失敗 → OPEN に戻る

### 面接での語り方

> 「ストリーク達成時のメール通知にサーキットブレーカーを実装しました。CLOSED→OPEN→HALF_OPENの3状態で、外部サービスの障害時にリクエストを遮断して自分のサービスを守ります。通知はトランザクション外で実行し、通知の失敗がチェックイン処理に影響しない設計にしています。MercoinではIstio/Envoyがインフラ層でやっていることを、仕組みを理解するために手で書きました」

---

## 面接での全体ストーリー

> 「Streekという個人開発の習慣トラッキングアプリで、金融システムの基本概念を実装しました。
>
> まずトランザクションマネージャーを作り、複数のDB操作を原子的に管理できるようにしました。次に台帳パターンでポイントシステムを実装し、残高を履歴の合計で算出する設計にしました。冪等性キーで二重処理も防止しています。ドメインイベントで処理を疎結合にし、将来のPub/Sub化に備えた設計にしました。さらに外部API連携（メール通知）にサーキットブレーカーを実装し、外部サービスの障害がアプリ本体に波及しない設計にしています。
>
> MercoinのSagaパターンや台帳管理とは規模が全く違いますが、金融システムの設計思想（原子性、監査性、冪等性、疎結合、障害分離）を小さく体験することで、概念だけでなく実装レベルで理解できました」
