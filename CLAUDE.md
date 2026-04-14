# CLAUDE.md - Streek

## Project Overview

**Streek** is a daily habit tracking app built with a Go API backend and SwiftUI frontend.
The primary purpose of this project is to demonstrate Go backend engineering skills for a backend engineer position at Mercari.

Architecture is inspired by patterns proven in production at Kalonade (100+ tenant SaaS),
adapted to a smaller scope without over-engineering.

---

## AI Usage Rules (CRITICAL)

### ❌ NO AI — Write everything by hand
- All Go code (handlers, middleware, models, repository, services, tests)
- Error handling design
- Directory structure decisions
- API design and routing
- Value Object / Entity design

### ✅ AI Allowed
- SwiftUI client code (all of it)
- Docker / docker-compose configuration
- PostgreSQL setup and SQL migration files
- README English proofreading
- CI/CD configuration (GitHub Actions)

### ⚠️ Gray Zone
- Go syntax questions → Use go.dev/doc and Go by Example, NOT AI
- Error debugging → Google the error message, Stack Overflow is OK, do NOT paste into AI
- Architecture/design discussion → Conceptual discussion with AI is OK (e.g., "How should JWT middleware be structured?"), but do NOT ask AI to generate code
- The rule: **"Can I explain this code line-by-line in an interview?"** If not, don't use it.

### Editor Setup
- **Go**: Zed (AI Assistant OFF) — install Go extension for gopls support
- **SwiftUI**: Xcode + Cursor for AI assistance

---

## Tech Stack

### Backend (Go) — The focus
- **Language**: Go 1.22+
- **Framework**: Echo v4
- **Database**: PostgreSQL
- **Cache**: Redis (streak/stats caching)
- **Auth**: JWT (self-implemented, no third-party auth library)
- **DB Access**: sqlx (lightweight, raw SQL friendly)
- **Migration**: golang-migrate
- **Validation**: go-ozzo/ozzo-validation (same approach as Kalonade Value Objects)
- **Error Handling**: cockroachdb/errors (stack trace付きエラー伝播)
- **Testing**: Go standard `testing` package + testify
- **Container**: Docker + docker-compose

### Frontend (SwiftUI) — Minimal effort
- **Platform**: iOS (SwiftUI)
- **Networking**: URLSession
- **State Management**: SwiftUI native (@State, @Observable)
- **Design**: Apple system defaults — do NOT spend time on custom UI

---

## Architecture (Clean Architecture — Kalonade Pattern Adapted)

Kalonadeでは5層 + アクター別ツリー分割を採用していたが、
Streekは単一ユーザー種別の小規模アプリのため、アクター分割は不要。
コア原則（依存逆転、レイヤー責務分離）だけを引き継ぐ。

```
streek/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point
│   ├── internal/
│   │   ├── handler/                 # Presentation層: Echo handlers（極薄）
│   │   │   ├── auth_handler.go
│   │   │   ├── habit_handler.go
│   │   │   ├── checkin_handler.go
│   │   │   └── stats_handler.go
│   │   ├── middleware/              # JWT auth middleware（context伝播）
│   │   ├── domain/                  # Domain層: Entity / Value Object / Repository Interface
│   │   │   ├── entity/
│   │   │   │   ├── user.go
│   │   │   │   ├── habit.go
│   │   │   │   └── checkin.go
│   │   │   ├── value_object/
│   │   │   │   ├── user_id.go
│   │   │   │   ├── habit_id.go
│   │   │   │   ├── email.go
│   │   │   │   ├── habit_name.go
│   │   │   │   └── hex_color.go
│   │   │   └── repository/         # Interface定義のみ（実装はinfrastructure）
│   │   │       ├── user_repository.go
│   │   │       ├── habit_repository.go
│   │   │       └── checkin_repository.go
│   │   ├── service/                 # Application層: 1ファイル1ユースケース
│   │   │   ├── register_user.go
│   │   │   ├── login_user.go
│   │   │   ├── create_habit.go
│   │   │   ├── update_habit.go
│   │   │   ├── delete_habit.go
│   │   │   ├── check_in.go
│   │   │   ├── undo_check_in.go
│   │   │   └── get_stats.go
│   │   ├── infrastructure/          # Infrastructure層: Repository実装 + 外部接続
│   │   │   ├── persistence/         # sqlx による Repository 実装
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── habit_repository.go
│   │   │   │   └── checkin_repository.go
│   │   │   ├── cache/               # Redis
│   │   │   │   └── streak_cache.go
│   │   │   └── auth/                # JWT token generation/validation
│   │   │       └── jwt.go
│   │   ├── config/                  # 環境設定
│   │   └── apperror/                # カスタムエラー型 + エラーコード体系
│   ├── migrations/                  # SQL migration files
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── ios/
│   └── Streek/                      # Xcode project
├── CLAUDE.md
└── README.md
```

### 依存方向（厳守）

```
Handler → Service → Domain ← Infrastructure
```

Infrastructure は Domain の Interface を実装することで依存逆転を成立させる。
Service 層が Infrastructure を直接 import することは禁止。

---

## Design Patterns (Kalonade由来)

### 1. Value Object — 生成時バリデーション、以降は常に妥当

Kalonadeのジェネリック基底 `LiteralBase[T]` を参考に、
Streekでは ozzo-validation でシンプルに実装する。

```go
// domain/value_object/email.go
type Email struct {
    value string
}

func NewEmail(s string) (*Email, error) {
    err := validation.Validate(s,
        validation.Required,
        is.EmailFormat,
        validation.RuneLength(1, 255),
    )
    if err != nil {
        return nil, err
    }
    return &Email{value: s}, nil
}

func (e Email) String() string { return e.value }
```

- 原始型(string, int)を直接使わない。必ず Value Object で包む
- Service層の Input には Value Object で受け取る（原始型で受けない）
- Optional は pointer (*Email)、nil チェック必須

### 2. Entity 更新規約 — Set*() メソッドで個別更新

Kalonadeで最も重要な規約。NewEntity() での全再構築は禁止。

```go
// ✅ 正しい: 個別フィールド更新
existing, _ := repo.FindByID(ctx, id)
existing.SetName(newName)
existing.SetUpdatedAt(time.Now())
repo.Save(ctx, existing)

// ❌ 禁止: フィールド渡し忘れで既存値が消えるバグの原因
entity := entity.NewHabit(existing.ID(), newName, nil, ...)
```

Entity 生成の使い分け:

| 関数 | 用途 |
|---|---|
| `CreateHabit()` | 新規作成時（Service層から呼ぶ） |
| `NewHabit()` | DB → Entity 復元時のみ（Repository内でのみ使用） |
| `Set*()` | 既存 Entity の更新時 |

### 3. Repository Interface はドメイン層に配置

```go
// domain/repository/habit_repository.go
type IHabitRepository interface {
    FindByID(ctx context.Context, id value_object.HabitID) (*entity.Habit, error)
    FindByUserID(ctx context.Context, userID value_object.UserID) ([]*entity.Habit, error)
    Save(ctx context.Context, habit entity.Habit) (*entity.Habit, error)
    Delete(ctx context.Context, id value_object.HabitID) error
}
```

実装は `infrastructure/persistence/` に置く。依存逆転の徹底。

### 4. 1ファイル1ユースケース（Service層）

Kalonadeの `Do()` メソッドパターンを踏襲。

```go
// service/create_habit.go
type CreateHabit struct {
    repo repository.IHabitRepository
}

type CreateHabitInput struct {
    UserID value_object.UserID
    Name   value_object.HabitName
    Color  value_object.HexColor
}

func NewCreateHabit(repo repository.IHabitRepository) *CreateHabit {
    return &CreateHabit{repo: repo}
}

func (s *CreateHabit) Do(ctx context.Context, input CreateHabitInput) (*entity.Habit, error) {
    habit := entity.CreateHabit(input.UserID, input.Name, input.Color)
    saved, err := s.repo.Save(ctx, habit)
    if err != nil {
        return nil, errors.Wrap(err, "failed to save habit")
    }
    return saved, nil
}
```

### 5. Handler は極薄 — 4ステップのみ

Kalonadeの Connect-go ハンドラと同じ思想を Echo で再現。

```
1. Request から値を取り出す
2. Value Object に変換（バリデーション）
3. Service.Do() を呼ぶ
4. Response を返す
```

ハンドラにビジネスロジックを書かない。

### 6. 認証情報は context で伝播

Kalonadeの `context.WithValue` パターンを踏襲。

```go
// middleware で JWT 検証後、context に UserID を注入
ctx = context.WithValue(ctx, userIDKey, userID)

// handler で取り出す
userID := ctx.Value(userIDKey).(value_object.UserID)
```

handler の引数で回さない。

### 7. カスタムエラー型 + エラーコード体系

Kalonadeの `MyError` + `ErrorCode` パターンを簡略化。

```go
// apperror/error.go
type AppError struct {
    Code    ErrorCode
    Status  int
    Message string
    Err     error
}

type ErrorCode string

const (
    CodeNotFound     ErrorCode = "NOT_FOUND"
    CodeBadRequest   ErrorCode = "BAD_REQUEST"
    CodeUnauthorized ErrorCode = "UNAUTHORIZED"
    CodeConflict     ErrorCode = "CONFLICT"
    CodeInternal     ErrorCode = "INTERNAL"
)

// handler で統一的にエラーレスポンスを返す
// { "error": { "code": "NOT_FOUND", "message": "habit not found" } }
```

エラーは `cockroachdb/errors.Wrap()` でスタック付き伝播。
Handler 層で `AppError` に変換してHTTPレスポンスにする。

### 8. テスト — テーブル駆動 + Interface Mock

```go
func TestCreateHabit_Do_Success(t *testing.T) {
    t.Parallel()
    repo := &mockHabitRepository{
        saveFunc: func(ctx context.Context, h entity.Habit) (*entity.Habit, error) {
            return &h, nil
        },
    }
    svc := NewCreateHabit(repo)
    result, err := svc.Do(ctx, CreateHabitInput{...})
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

- 命名: `Test{対象}_{メソッド}_{状況}`
- `t.Parallel()` 必須
- Repository は Interface mock に差替（DI の恩恵）
- Value Object は `Validate()` のケース網羅

---

## Kalonadeから意図的に省略したもの

| Kalonadeパターン | Streekでの判断 | 理由 |
|---|---|---|
| samber/do (DI) | 手動DI (main.goで組み立て) | 小規模で型安全DIコンテナは過剰 |
| Connect-go / protobuf | Echo REST + JSON | 単一クライアント(iOS)、gRPC不要 |
| アクター別ツリー分割 | 単一ユーザー種別 | staff/admin/user の区別が不要 |
| GORM | sqlx | 生SQL制御を重視、magic を避ける |
| マルチバイナリ (cmd/ 分割) | 単一API サーバー | workerが不要な規模 |
| Datadog APM / Sentry | なし | 個人プロジェクトでは不要 |
| TransactionManager (ctx埋込tx) | 必要になったら追加 | 現時点では単一DB操作が多い |

面接では「なぜ省略したか」も説明できるようにしておく。

---

## API Endpoints

### Auth
- `POST /api/v1/auth/register` — Create account (email + password)
- `POST /api/v1/auth/login` — Login, returns JWT access + refresh token
- `POST /api/v1/auth/refresh` — Refresh access token

### Habits
- `GET    /api/v1/habits` — List user's habits
- `POST   /api/v1/habits` — Create a habit
- `PUT    /api/v1/habits/:id` — Update a habit
- `DELETE /api/v1/habits/:id` — Delete a habit

### Check-ins
- `POST   /api/v1/habits/:id/check` — Mark habit as done today
- `DELETE /api/v1/habits/:id/check` — Undo today's check-in

### Stats
- `GET /api/v1/habits/:id/stats` — Streak count, completion rate, history
- `GET /api/v1/stats/overview` — All habits summary

---

## Data Models

### User
- id (UUID)
- email (unique)
- password_hash
- created_at

### Habit
- id (UUID)
- user_id (FK → User)
- name (varchar 50)
- description (nullable, varchar 200)
- color (varchar 7, hex e.g. #FF5733)
- created_at
- updated_at

### CheckIn
- id (UUID)
- habit_id (FK → Habit)
- checked_date (DATE, unique constraint per habit)
- created_at

---

## Key Implementation Details

### JWT Auth
- Access token: 15 min expiry
- Refresh token: 7 days expiry
- Password hashing: bcrypt
- Middleware extracts UserID from token → context.WithValue で注入

### Streak Calculation
- Current streak: count consecutive days backwards from today
- Longest streak: scan all check-ins and find max consecutive run
- Cache streaks in Redis, invalidate on check-in/undo

### Error Response Format
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "habit not found"
  }
}
```

---

## Development Steps

### Step 1: Go API Core (Day 1-2)
1. Project setup (go mod init, Echo, sqlx, Docker)
2. Database schema + migrations
3. Value Objects (UserID, HabitID, Email, HabitName, HexColor)
4. User entity + repository interface + sqlx implementation
5. Register & Login with JWT (self-implemented)
6. Auth middleware (context伝播)

### Step 2: Habit CRUD + Check-ins (Day 2-3)
1. Habit entity (CreateHabit / NewHabit / Set* pattern)
2. Habit repository interface + implementation
3. Service layer (1 file per use case, Do() method)
4. CheckIn entity + repository + service
5. Input validation via Value Objects
6. Unit tests (table-driven + mock)

### Step 3: Stats & Redis (Day 3-4)
1. Streak calculation logic (service layer)
2. Completion rate calculation
3. Redis caching for streaks
4. Integration tests

### Step 4: SwiftUI Client (Day 4-5) — AI OK
1. Login / Register screens
2. Habit list with check-in toggles
3. Simple stats view

### Step 5: Polish (Day 5-6)
1. README.md (English, architecture diagram, setup instructions)
2. docker-compose one-command startup
3. GitHub repo cleanup

---

## Interview Talking Points (Build These While Coding)

### Architecture
- Why Clean Architecture? How does dependency inversion work in Go?
- How did you separate domain logic from infrastructure?
- Why Repository Interface in domain layer, not in infrastructure?
- What did you adopt from your production codebase? What did you simplify, and why?

### Go Specifics
- How did you implement JWT auth without a library?
- How does the context-based auth propagation work?
- How did you design Value Objects in Go?
- Error handling: why cockroachdb/errors? How do you convert to HTTP responses?
- Why sqlx over GORM? (raw SQL control, no magic, explicit queries)

### Design Decisions
- Why Entity update via Set*() instead of reconstruction? What bug does it prevent?
- Why 1-file-1-usecase in the service layer?
- How would you add transaction management if needed? (context埋込tx pattern)
- What patterns from your production SaaS did you intentionally omit, and why?

### Scale Questions
- What would you change if this needed to scale to 1M users?
- How would you split this into microservices?
- How would you add async processing (e.g., notification worker)?
- How would you add rate limiting to the API?
