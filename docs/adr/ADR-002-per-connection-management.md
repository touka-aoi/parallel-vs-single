# ADR: Per-Connection Owner Loop による接続管理設計

# Status
- Draft: 記述中またはレビュー中

# Decision
１接続ごとにI/0・状態管理・アプリケーション層を担当するループ（goroutine）を設ける設計とする。

### 概念整理
`Conn` は 物理的な接続 を表す
- WebSocket の実体
- 切断されれば再利用されない
`Session` は 論理的な接続 を表す
- アプリケーションから見た接続単位
- 将来的に再接続による rebind を想定
### ループ構成（per-conn / per-session）
owner loop
- 状態管理・判断の唯一の場所
- 切断、timeout、heartbeat、sendQ の扱いを決定する

readLoop
- I/O Read を担当（ブロッキング）
- 読み取ったデータは アプリケーションのイベントループ（Room/dispatcher）へ通知する
- owner loop には 接続管理に必要な制御イベント（read touch / pong / I/O error / close request）のみ通知する

writeLoop
- I/O Write を担当
- 書き込み結果や I/O エラーをイベントとして owner loop に通知

heartbeatLoop
- ticker による定期的な tick を owner loop に送信
- 状態更新や切断判断は行わない

# Context
本システムは WebSocket を用いたリアルタイム通信を前提とし、多数の同時接続を扱う。
接続管理において以下を考慮する必要があった。
- read / write / heartbeat といった複数の goroutine から同一の接続・状態を操作することによる競合リスク
- 切断、timeout、heartbeat、backpressure などの判断が分散し、責務が不明瞭になる問題
- アプリケーション（Room）層に接続管理の詳細が漏れ出す設計上の違和感
- 将来的な再接続（論理セッションの継続）を見据えた拡張性の不足
これらを踏まえ、「I/O・状態管理・アプリケーション層構成が分離・強調できる構成」が求められた。

# Consideration
1. 状態更新を atomic にして複数 goroutine から直接操作する
   - 実装は簡単だが、 状態遷移が分散 close / timeout の判断が複数箇所に分かれる
   - 設計が成長すると破綻しやすいため不採用
2. 全接続を 1 つのグローバル event loop で管理する
   - goroutine の削減が可能
   - websocket実装ではper-connection Read のブロッキングが避けられないため、今回は見送り
   - 他プロトコルの採用時に再検討の余地あり

# Consequences
Pros
- 切断、timeout、heartbeat、backpressure の判断が 1 箇所に集約される
- race condition を設計レベルで排除できる
- Conn（物理）と Session（論理）の責務分離が明確
- 将来的な再接続（Session rebind）に拡張しやすい
- Room / アプリケーション層がシンプルになる
Cons
- 接続ごとに goroutine 数が増える（read/write/heartbeat + owner）

# References
参考情報
