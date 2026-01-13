package domain

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type RoomID string

var ErrRoomBusy = errors.New("room control channel is full")

type Room struct {
	ID       RoomID
	sessions map[SessionID]struct{}

	dispatcher  Dispatcher
	application Application // 外部からアプリケーションロジックを注入できる

	ctrlCh    chan roomCtrl
	receiveCh chan []byte
	sendCh    chan roomSend

	tickInterval time.Duration
}

func NewRoom(id RoomID) *Room {
	return &Room{
		ID:           id,
		sessions:     make(map[SessionID]struct{}),
		ctrlCh:       make(chan roomCtrl, 1024),
		receiveCh:    make(chan []byte, 1024),
		sendCh:       make(chan roomSend, 1024),
		tickInterval: time.Second / 60,
	}
}

func (r *Room) Receive(ctx context.Context, data []byte) error {
	select {
	case <-ctx.Done():
		return nil
	case r.receiveCh <- data:
		return nil
	default:
		return ErrRoomBusy
	}
}

// AddRoom はルームに接続を追加します。チャネルが満杯の場合、ErrRoomBusy を返します。
func (r *Room) AddRoom(ctx context.Context, sessionID SessionID) error {
	select {
	case <-ctx.Done():
		return nil
	case r.ctrlCh <- roomCtrl{kind: roomCtrlAdd, sessionID: sessionID}:
		return nil
	default:
		return ErrRoomBusy
	}
}

// RemoveRoom はルームから接続を削除します。チャネルが満杯の場合、ErrRoomBusy を返します。
func (r *Room) RemoveRoom(ctx context.Context, sessionID SessionID) error {
	select {
	case <-ctx.Done():
		return nil
	case r.ctrlCh <- roomCtrl{kind: roomCtrlRemove, sessionID: sessionID}:
		return nil
	default:
		return ErrRoomBusy
	}
}

func (r *Room) Broadcast(ctx context.Context, data []byte) error {
	for sessionID := range r.sessions {
		if err := r.dispatcher.Dispatch(ctx, sessionID, data); err != nil {
			return err
		}
	}
	return nil
}

func (r *Room) SendTo(ctx context.Context, sessionID SessionID, data []byte) error {
	return r.dispatcher.Dispatch(ctx, sessionID, data)
}

func (r *Room) EnqueueBroadcast(ctx context.Context, data []byte) error {
	return r.enqueueSend(ctx, roomSend{kind: roomSendBroadcast, data: data})
}

func (r *Room) EnqueueSendTo(ctx context.Context, sessionID SessionID, data []byte) error {
	return r.enqueueSend(ctx, roomSend{kind: roomSendTo, sessionID: sessionID, data: data})
}

func (r *Room) enqueueSend(ctx context.Context, msg roomSend) error {
	select {
	case <-ctx.Done():
		return nil
	case r.sendCh <- msg:
		return nil
	default:
		return ErrRoomBusy
	}
}

func (r *Room) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.tickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		CTRL_LOOP:
			for {
				select {
				case ctrl := <-r.ctrlCh:
					r.handleControlMessage(ctrl)
				default:
					break CTRL_LOOP
				}
			}
		RECEIVE_LOOP:
			for {
				select {
				case data := <-r.receiveCh:
					// アプリケーションロジックが担当する
					parseData, err := r.application.Parse(ctx, data)
					if err != nil {
						// どうするのこれ？ なんでルームでパースしてるかは意味わからんすぎるな
					}
					if err := r.application.Handle(ctx, parseData); err != nil {
						// どうするのこれ？
					}
				default:
					break RECEIVE_LOOP
				}
			}
			// 送信するデータがあれば送信する このデータは１フレーム前のデータになる
		SEND_LOOP:
			for {
				select {
				case msg := <-r.sendCh: // アプリケーションが１tick前に処理したデータがここに入っている
					r.handleSendMessage(ctx, msg)
				default:
					break SEND_LOOP
				}
			}
		}
	}
}

func (r *Room) handleControlMessage(ctrl roomCtrl) {
	switch ctrl.kind {
	case roomCtrlAdd:
		r.sessions[ctrl.sessionID] = struct{}{}
	case roomCtrlRemove:
		delete(r.sessions, ctrl.sessionID)
	default:
	}
}

func (r *Room) handleSendMessage(ctx context.Context, msg roomSend) {
	switch msg.kind {
	case roomSendBroadcast:
		if err := r.Broadcast(ctx, msg.data); err != nil {
			slog.WarnContext(ctx, "room broadcast failed", "err", err)
		}
	case roomSendTo:
		if err := r.SendTo(ctx, msg.sessionID, msg.data); err != nil {
			slog.WarnContext(ctx, "room send failed", "err", err, "session_id", msg.sessionID)
		}
	default:
	}
}
