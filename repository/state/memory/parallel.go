package memory

import (
	"context"
	"sync"
	"time"

	"github.com/touka-aoi/paralle-vs-single/domain"
	"github.com/touka-aoi/paralle-vs-single/repository/state"
)

type ConcurrentStore struct {
	base *Store
	mu   sync.RWMutex
}

func NewConcurrentStore() *ConcurrentStore {

	return &ConcurrentStore{
		base: newStore(),
	}
}

func (c *ConcurrentStore) ApplyMove(ctx context.Context, cmd *state.Move) (*domain.MoveResult, error) {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.base.applyMove(cmd, time.Now())
}

func (c *ConcurrentStore) ApplyAttack(ctx context.Context, cmd *state.Attack) (*domain.AttackResult, error) {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.base.applyAttack(cmd, time.Now())
}

func (c *ConcurrentStore) RegisterPlayer(ctx context.Context, playerID string, roomID string) error {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.base.registerPlayer(playerID, roomID)
}

var _ state.InteractionState = (*ConcurrentStore)(nil)
