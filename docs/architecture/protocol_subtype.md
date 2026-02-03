# SubType設計

## 概要

`PayloadHeader.SubType`は`uint8`として定義。`DataType`によって解釈が変わる。

## SubTypeの解釈

| dataType | subTypeの解釈 |
|----------|---------------|
| actor (2) | ActorSubType (spawn=1, update=2, despawn=3) |
| control (4) | ControlSubType (join=1, leave=2, kick=3, ping=4, pong=5, error=6) |

## 実装

```go
type PayloadHeader struct {
    DataType DataType
    SubType  uint8  // DataTypeに応じてキャストする
}

// 使用例
switch ph.DataType {
case DataTypeActor:
    actorSubType := ActorSubType(ph.SubType)
case DataTypeControl:
    controlSubType := ControlSubType(ph.SubType)
}
```