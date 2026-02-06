# セッションID通知機能の実装プラン

## 概要
サーバーからクライアントへ接続時にセッションIDを通知する機能を追加し、
クライアントが正しいセッションIDでJoinメッセージを送信できるようにする。

## 現状の問題
- クライアントは接続時に自分のセッションIDを知らない
- `onConnect`でJoinメッセージを送信するが、sessionID=0を使用
- サーバー側で`session ID mismatch`エラーが発生

## 変更後のフロー
```
Client                          Server
   |                               |
   |-------- WS Connect ---------> |
   |                               | SessionEndpoint作成 (ID=1)
   |<-- Control/Assign(ID=1) ----- |
   | mySessionId = 1               |
   |---- Control/Join(ID=1) -----> |
   |                               | ルームに参加
```

## 変更ファイル

### 1. server/domain/protocol.go
- `ControlSubTypeAssign = 7` を追加
- `EncodeAssignMessage(sessionID SessionID) []byte` 関数を追加
  - Header + PayloadHeader (DataType=Control, SubType=Assign) を構築

### 2. server/domain/session_endpoint.go
- `Run()` の最初（goroutine起動前）でセッションID通知メッセージを送信
  ```go
  func (se *SessionEndpoint) Run() error {
      // セッションID通知を送信
      assignMsg := EncodeAssignMessage(se.session.ID())
      if err := se.connection.Write(se.ctx, assignMsg); err != nil {
          return err
      }
      // ...既存の処理
  }
  ```

### 3. client/src/protocol.ts
- `CONTROL_SUBTYPE_ASSIGN = 7` を追加
- `getControlSubType(data: ArrayBuffer): number` 関数を追加
- `decodeAssignMessage(data: ArrayBuffer): number` 関数を追加
  - HeaderからsessionIdを取得して返す

### 4. client/src/game.ts
- `onConnect()` を修正: Joinメッセージを送信しない（待機のみ）
- `onMessage()` を修正:
  - `DATA_TYPE_CONTROL` かつ `CONTROL_SUBTYPE_ASSIGN` の場合:
    - `decodeAssignMessage`でsessionIdを取得
    - `mySessionId`に保存
    - Joinメッセージを正しいsessionIdで送信

## 実装順序
1. protocol.go: 定数と`EncodeAssignMessage`を追加
2. session_endpoint.go: `Run()`でセッションID通知を送信
3. protocol.ts: 定数と`decodeAssignMessage`を追加
4. game.ts: onConnect/onMessageのロジック変更

## 検証方法
1. サーバー起動: `go run server/cmd/main.go`
2. クライアント起動: `cd client && npm run dev`
3. ブラウザでアクセスし、以下を確認:
   - コンソールに「session ID mismatch」が出ないこと
   - 「session joined room」のログが出ること
   - ゲームが正常に動作すること
