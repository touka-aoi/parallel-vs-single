ADR: SessionID.Bytes()はエラーを返さない

# Status
Accepted

# Decision
`SessionID.Bytes()`メソッドはエラーを返さず、`[16]byte`のみを返す。
内部でbase64デコードエラーが発生した場合はゼロ値を返す。

# Context
`SessionID`は内部的にbase64エンコードされた16バイトのランダム値を文字列として保持している。
`Bytes()`メソッドはこの文字列をbase64デコードして`[16]byte`に変換する。

当初、`base64.RawURLEncoding.DecodeString`がエラーを返す可能性があるため、
`Bytes()`もエラーを返すように設計されていた。

しかし、すべての呼び出し箇所でエラーハンドリングが必要になり、コードが冗長になっていた。

# Consideration
**前提条件:**
- `SessionID`は常に`NewSessionID()`で生成される内部型である
- 外部からの入力で`SessionID`を直接構築することはない
- したがって、正しくbase64エンコードされた値のみが存在する

**リスク:**
- 将来、外部入力から`SessionID`を構築する場合、デコードエラーがサイレントに無視される
- テストで不正な`SessionID`を使用した場合、予期しない動作をする可能性がある

**代替案:**
1. エラーを返すまま維持 → 呼び出し側が冗長になる
2. `MustBytes()`を追加しpanicする → 本番環境でのpanicリスク
3. エラーを握りつぶしゼロ値を返す → シンプルだが、前提条件が必要

**結論:**
現在の設計では`SessionID`は内部でのみ生成されるため、前提条件を明記した上で
オプション3を採用する。

# Consequences
- 呼び出し側のコードがシンプルになる
- 将来`SessionID`の構築方法を変更する場合は、この決定を再検討する必要がある
- コメントに前提条件を明記し、将来の開発者に意図を伝える

# References
- `server/domain/session.go`: `SessionID.Bytes()`の実装