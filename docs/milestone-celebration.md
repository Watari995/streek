# Streak Milestone Celebration (iOS)

## 背景

バックエンドは check-in 時の `CheckInSucceededEvent` を受けて、ストリークが
マイルストーンに到達すると **メール通知** を送る
（[notify_streak_milestone.go](../backend/internal/application/eventhandler/notify_streak_milestone.go)、
しきい値は [streak_service.go](../backend/internal/domain/service/streak_service.go) の
`streakMilestones = [10,20,30,40,50,60,70,80,90]`）。

アプリはこのメールを受け取れないため、ユーザーはマイルストーン到達を体験できなかった。
追加のバックエンドAPIは作らず、**既存の stats overview が返す `current_streak`** を
使ってアプリ内で祝福演出を出すようにした（アプリ側だけで完結）。

## 実装した箇所

| ファイル | 内容 |
|---|---|
| [MilestoneCelebrationView.swift](../ios/Streek/Streek/Features/Habits/MilestoneCelebrationView.swift) | 新規。`StreekMilestone` 判定ロジック + 祝福オーバーレイUI |
| [HabitListView.swift](../ios/Streek/Streek/Features/Habits/HabitListView.swift) | check-in 後にマイルストーン到達を検出して演出を表示 |

### 1. `StreakMilestone`（新規ファイル内）

- `thresholds = [10,20,30,40,50,60,70,80,90]` — バックエンドの `streakMilestones`
  と一字一句一致させている（コメントで同期義務を明記）。
- `reached(before:after:) -> Int?` — check-in 前後のストリークを比較し、
  `after > before` かつ `after` がしきい値ちょうどのときだけマイルストーン値を返す。
  - undo（`after <= before`）では発火しない
  - アプリ再起動や再描画では `before/after` が変化しないため再発火しない

### 2. `MilestoneCelebrationView`（新規ファイル内）

- 全画面の半透明オーバーレイ + 中央カード（炎アイコン / `N-day streak!` / habit名）。
- spring アニメーションで出現、`UINotificationFeedbackGenerator(.success)` で触覚。
- **2.8秒で自動消滅** + タップでも消える（操作をブロックしない）。
- 既存の `Theme.swift` トークン（`appAccent` / `appSurface` / `AppFont` 等）のみ使用。

### 3. `HabitListView` への配線

- `Celebration`（`Identifiable`）state を追加。
- `toggle(_:)` で **check-in 前の `current_streak` を `streakBefore` に退避** →
  `checkInStore.toggle` 成功後に `statsStore.loadOverview()` を **await**
  （従来は fire-and-forget だったが、マイルストーン判定のため待つように変更。
  points 残高更新は従来どおり並行のまま）→ `streakAfter` と比較し、
  `StreekMilestone.reached` が値を返したら `celebration` をセット。
- `.overlay` で `celebration` 非nil時に `MilestoneCelebrationView` を表示。

## やらなかったこと（スコープ外）

バックエンド実装が必要なため今回は未対応：

- **ポイント SPEND/REDEEM**：アプリの表示パス（[PointsView.swift](../ios/Streek/Streek/Features/Points/PointsView.swift)）は
  既に SPEND/REDEEM 対応済みだが、ポイントを使う/交換するエンドポイントが
  バックエンドに無いため死にコードのまま。
- **リフレッシュトークン** (`POST /api/v1/auth/refresh`)：バックエンド未実装。
- **habit個別 stats** (`GET /api/v1/habits/:id/stats`)、check-in履歴一覧API：バックエンド未実装。

これらは Go 側の実装が必要で、CLAUDE.md の AI 利用ルール上、別途明示指示が要る。

## 検証

`xcodebuild build -scheme Streek`（iPhone 17 Pro / iOS 26.4 simulator）→ **BUILD SUCCEEDED**。
