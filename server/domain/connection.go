package domain

import "context"

type ConnectionID string

// Connection は物理的な接続を表します。
type Connection struct {
	ID        ConnectionID
	transport Transport
}

func NewConnection(transport Transport) *Connection {
	return &Connection{transport: transport}
}

func (c *Connection) Write(ctx context.Context, data []byte) error {
	return c.transport.Write(ctx, data)
}

func (c *Connection) Read(ctx context.Context) ([]byte, error) {
	return c.transport.Read(ctx)
}

func (c *Connection) Close() {
	_ = c.transport.Close(1000, "")
}
