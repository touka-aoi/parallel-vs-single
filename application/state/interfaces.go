package state

import (
	"context"
	"time"

	"github.com/touka-aoi/paralle-vs-single/application/domain"
)

// InteractionState はサービス層から呼び出される状態更新インターフェース。
// デフォルトのインメモリ実装は `application/state/memory` パッケージを参照。
// 並列実装・単一ループ実装はこの契約に従ってドメインコマンドを適用する。
type InteractionState interface {
	ApplyMove(ctx context.Context, cmd domain.MoveCommand) (domain.MoveResult, error)
	ApplyBuff(ctx context.Context, cmd domain.BuffCommand) (domain.BuffResult, error)
	ApplyAttack(ctx context.Context, cmd domain.AttackCommand) (domain.AttackResult, error)
	ApplyTrade(ctx context.Context, cmd domain.TradeCommand) (domain.TradeResult, error)
}

// MetricsRecorder は各実装の統計収集を抽象化する。
type MetricsRecorder interface {
	RecordLatency(ctx context.Context, endpoint string, duration time.Duration)
	RecordContention(ctx context.Context, endpoint string, wait time.Duration)
	IncrementCounter(ctx context.Context, name string, delta int)
}
