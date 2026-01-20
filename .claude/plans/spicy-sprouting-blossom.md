# Pub/Sub パターン導入計画

## 概要

現在の Dispatcher/Sender パターンを Pub/Sub パターンに置き換え、SessionEndpoint と Room の疎結合化を実現する。

## 現状の問題

1. **Dispatcherの責務が不明確**: 受信配送と送信配送が混在
2. **SessionIDの伝播がない**: `Dispatch(ctx, data)` にSessionIDがない
3. **Senderの設計が複雑**: SessionRegistry等の追加検討が必要だった

## 解決策: Pub/Sub パターン

### メッセージフロー

```
【受信方向: Client → Room】
Client → SessionEndpoint.readLoop()
       → pubsub.Publish("room:{roomID}", data)
       → Room.Subscribe("room:{roomID}")
       → Room で処理

【送信方向: Room → Client】
Room → pubsub.Publish("session:{sessionID}", data)
     → SessionEndpoint.Subscribe("session:{sessionID}")
     → writeLoop() → Client
```

## 変更ファイル一覧

### 新規作成

| ファイル | 内容 |
|---------|------|
| `server/domain/pubsub.go` | PubSub インターフェース定義 |
| `server/domain/simple_pubsub.go` | インメモリ PubSub 実装 |
| `server/domain/room_manager.go` | RoomManager インターフェース定義 |
| `server/domain/simple_room_manager.go` | 固定ルームを返すシンプルな実装 |

### 修正

| ファイル | 変更内容 |
|---------|----------|
| `server/domain/session_endpoint.go` | Dispatcher → PubSub に変更、subscribe/publish追加 |
| `server/domain/room.go` | Dispatcher → PubSub に変更 |
| `server/domain/dispatcher.go` | 削除または非推奨化 |
| `server/handler/accept.go` | PubSub を渡すように変更 |
| `server/router.go` | PubSub を渡すように変更 |
| `server/cmd/main.go` | PubSub 初期化を追加 |

## 実装詳細

### 1. PubSub インターフェース (`server/domain/pubsub.go`)

```go
package domain

type PubSub interface {
    // Subscribe はトピックを購読し、メッセージを受信するチャネルを返す
    Subscribe(topic string) <-chan Message

    // Unsubscribe は購読を解除する
    Unsubscribe(topic string, ch <-chan Message)

    // Publish はトピックにメッセージを配信する
    Publish(topic string, data []byte) error
}

type Message struct {
    Topic string
    Data  []byte
}
```

### 2. SimplePubSub 実装 (`server/domain/simple_pubsub.go`)

```go
package domain

type SimplePubSub struct {
    mu          sync.RWMutex
    subscribers map[string][]chan Message
}

func NewSimplePubSub() *SimplePubSub { ... }
func (p *SimplePubSub) Subscribe(topic string) <-chan Message { ... }
func (p *SimplePubSub) Unsubscribe(topic string, ch <-chan Message) { ... }
func (p *SimplePubSub) Publish(topic string, data []byte) error { ... }
```

### 3. RoomManager インターフェース (`server/domain/room_manager.go`)

RoomManagerをインターフェースにしておくことで、将来マッチングツール等を作成した際に実装を差し替え、マッチングサービスに問い合わせる形に拡張できる。

```go
package domain

import "context"

// RoomManager はセッションに対するルーム割り当てを管理する
// 将来的にマッチングサービスへの問い合わせ等に差し替え可能
type RoomManager interface {
    // GetRoom はセッションに割り当てるルームIDを返す
    GetRoom(ctx context.Context, sessionID SessionID) (RoomID, error)
}
```

### 4. SimpleRoomManager 実装 (`server/domain/simple_room_manager.go`)

```go
package domain

import "context"

// SimpleRoomManager は常に固定のルームを返すシンプルな実装
type SimpleRoomManager struct {
    defaultRoomID RoomID
}

func NewSimpleRoomManager(defaultRoomID RoomID) *SimpleRoomManager {
    return &SimpleRoomManager{defaultRoomID: defaultRoomID}
}

func (m *SimpleRoomManager) GetRoom(ctx context.Context, sessionID SessionID) (RoomID, error) {
    return m.defaultRoomID, nil
}
```

### 5. SessionEndpoint 修正

```go
type SessionEndpoint struct {
    session     *Session
    connection  *Connection
    pubsub      PubSub
    roomManager RoomManager  // RoomManager経由でルーム取得
    roomID      RoomID       // 実行時に決定

    ctrlCh  chan endpointEvent
    writeCh chan []byte
    // ...
}

func NewSessionEndpoint(session *Session, connection *Connection, pubsub PubSub, roomManager RoomManager) (*SessionEndpoint, error) {
    // ...
}

func (se *SessionEndpoint) Run() error {
    // RoomManagerにルームを問い合わせ
    roomID, err := se.roomManager.GetRoom(se.ctx, se.session.ID())
    if err != nil {
        return err
    }
    se.roomID = roomID

    // 自分宛のメッセージを購読
    sessionTopic := "session:" + se.session.ID().String()
    msgCh := se.pubsub.Subscribe(sessionTopic)
    defer se.pubsub.Unsubscribe(sessionTopic, msgCh)

    // room側にセッション追加を通知
    se.pubsub.Publish("room:"+string(se.roomID)+":ctrl", joinEvent{sessionID: se.session.ID()})
    defer se.pubsub.Publish("room:"+string(se.roomID)+":ctrl", leaveEvent{sessionID: se.session.ID()})

    // goroutine起動...
}

func (se *SessionEndpoint) readLoop(ctx context.Context) {
    for {
        data, err := se.connection.Read(ctx)
        // ...
        // roomにpublish（sessionIDを含める）
        se.pubsub.Publish("room:"+string(se.roomID), Message{
            SessionID: se.session.ID(),
            Data:      data,
        })
    }
}

// subscribeLoop: pubsubからのメッセージをwriteChに転送
func (se *SessionEndpoint) subscribeLoop(ctx context.Context, msgCh <-chan Message) {
    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-msgCh:
            se.writeCh <- msg.Data
        }
    }
}
```

### 6. Room 修正

```go
type Room struct {
    ID          RoomID
    sessions    map[SessionID]struct{}
    pubsub      PubSub      // dispatcher を置き換え
    application Application
    // ...
}

func (r *Room) Run(ctx context.Context) error {
    // room宛のメッセージを購読
    topic := "room:" + string(r.ID)
    msgCh := r.pubsub.Subscribe(topic)
    defer r.pubsub.Unsubscribe(topic, msgCh)

    for {
        select {
        case msg := <-msgCh:
            // アプリケーションロジックで処理
            result := r.application.Handle(ctx, msg.Data)
            // 結果を各セッションに配信
            for sessionID := range r.sessions {
                r.pubsub.Publish("session:"+sessionID.String(), result)
            }
        }
    }
}
```

## 実装順序

1. `domain/pubsub.go` - PubSubインターフェース定義
2. `domain/simple_pubsub.go` - PubSub実装
3. `domain/room_manager.go` - RoomManagerインターフェース定義
4. `domain/simple_room_manager.go` - 固定ルームを返す実装
5. `domain/session_endpoint.go` - PubSub + RoomManager対応
6. `domain/room.go` - PubSub対応
7. `handler/accept.go` - PubSub, RoomManagerを渡すように変更
8. `cmd/main.go` - PubSub, RoomManager, Room の初期化
9. 既存のDispatcher関連ファイルの整理（削除 or 非推奨化）

## 検証方法

1. `go build ./...` - ビルド確認
2. `go test ./server/domain/...` - 既存テストが通ること
3. サーバー起動 + WebSocket接続テスト
   - `go run server/cmd/main.go`
   - wscat等で接続し、メッセージ送受信を確認

## 決定事項

- **Room割り当て**: サーバー側で自動割り当て
  - 当面は単一Roomで運用
  - サーバー起動時にデフォルトRoomを1つ作成
  - 全ての接続はこのRoomに自動的に割り当て
  - 将来的に複数Room対応できる設計にしておく（RoomManager経由）
- **エラーハンドリング**: エラーを返す（現状維持）
  - `ErrBackpressure`等を返し、呼び出し側で対処
- **バッファサイズ**: 固定値（1024）
  - 現在のwriteChと同じサイズ
