# ADR SessionEndpoint の Close 設計
# Status: accepted

# Decision
1. 制御メッセージ（ctrlCh）送信はブロックを許容するが、呼び出しがわがタイムアウト等で制御できるようにする
2. 即時停止が必要な場合のため ForceClose() を提供する
3. Close 処理は一度のみ実行されるようにする

# Context
SessionEndpoint は 1 セッションに紐づく接続を管理し、ownerLoop/readLoop/writeLoop の複数 goroutine で動作する。
この構成では Close が複数経路（外部呼び出し・idle・Read/Write エラー等）から発生しうるため、以下の課題があった。
1. 制御メッセージ（Close 等）によってセッションは制御されるため、確実に ownerLoop に届く必要がある
2. 制御メッセージの送信が詰まってしまい長時間ブロッキングしてしまう可能性がある
3. ownerLoop が詰まってしまった場合、必要に応じて強制的に停止できるないといけない
4. Close が複数経路から呼ばれた場合に安全に動作する必要がある
