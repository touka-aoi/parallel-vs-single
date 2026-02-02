# クライアント主導のJoinRoom設計

## 背景
- 現状: `session_endpoint.go`が接続時に自動で"join"文字列を`ctrlTopic`に送信
- 問題: `Room.handleControlMessage`がセッションを追加するが、`Application.HandleMessage`が呼ばれずアクターがスポーンされない

## 新設計
マッチングは事前に完了している前提で、クライアントが接続確認後に明示的にControl/Joinメッセージを送信する。

### Join フロー
```
Client --[Control/Join]--> WebSocket
    --> session_endpoint.readLoop
    --> roomTopic (PubSub)
    --> room.Run RECEIVE_LOOP
    --> r.sessions に追加 (Control/Joinの場合)
    --> application.HandleMessage
    --> field.SpawnAtCenter (アクタースポーン)
```

### Leave フロー
1. クライアント: 切断前にControl/Leaveを送信（ベストエフォート）
2. サーバー: タイムアウトでセッション終了を検知 → Roomに通知

## 修正内容

### 1. server/domain/session_endpoint.go
- 自動join送信を削除 (L82-85)
- deferのleave送信も削除
- closeメソッドでControl/LeaveをroomTopicに送信（タイムアウト時も対応）

### 2. server/domain/room.go
- `ctrlTopic`購読を削除 (L74-77)
- `ctrlCh`関連のループを削除 (L87-96)
- `handleControlMessage`関数を削除 (L130-140)
- RECEIVE_LOOPでメッセージを検査し、Control/Join/Leaveなら`r.sessions`を更新してから`HandleMessage`を呼ぶ

### 3. client/src/protocol.ts
- `encodeControlMessage(subType: number)`関数を追加
- Control SubType定数を追加: `CONTROL_SUBTYPE_JOIN = 1`, `CONTROL_SUBTYPE_LEAVE = 2`

### 4. client/src/game.ts
- `onConnect`でControl/Joinメッセージを送信
- `destroy`でControl/Leaveメッセージを送信（ベストエフォート）

## 検証方法
1. `make server`でサーバー起動
2. `make client`でクライアント起動
3. ブラウザで接続 → サーバーログに`handleControl:join`が出力されることを確認
4. アクターが画面中央に表示されることを確認
5. WASDで移動できることを確認
6. ブラウザタブを閉じる → サーバーログに`handleControl:leave`が出力されることを確認