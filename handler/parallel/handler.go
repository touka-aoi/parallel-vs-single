package parallel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"

	"github.com/touka-aoi/paralle-vs-single/handler"
	"github.com/touka-aoi/paralle-vs-single/service"
)

const (
	scopeAck       = "ack"
	scopeBroadcast = "broadcast"
)

// Handler は WebSocket 経由で受信したフレームをサービス層へ橋渡しし、
// ハイインタラクションなイベントはブロードキャストで全クライアントへ配信する。
type Handler struct {
	svc     *service.InteractionService
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

// wsClient は送受信を独立ループで処理するための接続ラッパー。
type wsClient struct {
	conn *websocket.Conn
	send chan outboundFrame
}

// NewHandler は依存するサービスを受け取り、WebSocket ハンドラを構築する。
func NewHandler(svc *service.InteractionService) *Handler {
	return &Handler{
		svc:     svc,
		clients: make(map[*wsClient]struct{}),
	}
}

// ServeHTTP は WebSocket へアップグレードし、受信・送信ループを分離して開始する。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("parallel ws: accept error: %v", err)
		return
	}
	client := &wsClient{
		conn: conn,
		send: make(chan outboundFrame, 64),
	}
	h.addClient(client)
	ctx := r.Context()
	go h.writeLoop(ctx, client)
	h.readLoop(ctx, client)
}

func (h *Handler) readLoop(ctx context.Context, client *wsClient) {
	defer func() {
		h.removeClient(client)
		client.conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		msgType, data, err := client.conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
				log.Printf("parallel ws: read error: %v", err)
			}
			return
		}
		if msgType != websocket.MessageText {
			h.sendToClient(client, outboundFrame{Scope: scopeAck, Error: "invalid message type"})
			continue
		}
		resp, broadcasts := h.handleFrame(ctx, data)
		h.sendToClient(client, resp)
		if len(broadcasts) > 0 {
			h.broadcast(broadcasts)
		}
	}
}

func (h *Handler) writeLoop(ctx context.Context, client *wsClient) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("parallel ws: marshal error: %v", err)
				continue
			}
			if err := client.conn.Write(ctx, websocket.MessageText, data); err != nil {
				log.Printf("parallel ws: write error: %v", err)
				return
			}
		}
	}
}

func (h *Handler) addClient(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
}

func (h *Handler) removeClient(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
	close(client.send)
}

func (h *Handler) sendToClient(client *wsClient, frame outboundFrame) {
	select {
	case client.send <- frame:
	default:
		log.Printf("parallel ws: dropping message to client (buffer full)")
	}
}

func (h *Handler) broadcast(frames []outboundFrame) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		for _, frame := range frames {
			select {
			case client.send <- frame:
			default:
				log.Printf("parallel ws: dropping broadcast to client (buffer full)")
			}
		}
	}
}

type inboundFrame struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type outboundFrame struct {
	Type   string      `json:"type"`
	Scope  string      `json:"scope"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func (h *Handler) handleFrame(ctx context.Context, data []byte) (outboundFrame, []outboundFrame) {
	var frame inboundFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		return outboundFrame{Scope: scopeAck, Error: fmt.Sprintf("invalid frame: %v", err)}, nil
	}
	frameType := strings.ToLower(frame.Type)
	switch frameType {
	case "move":
		var payload handler.MovePayload
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return outboundFrame{Type: frameType, Scope: scopeAck, Error: fmt.Sprintf("invalid payload: %v", err)}, nil
		}
		result, err := h.svc.Move(ctx, &payload)
		return h.makeResponse(frameType, result, err, true)
	case "buff":
		var payload handler.BuffPayload
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return outboundFrame{Type: frameType, Scope: scopeAck, Error: fmt.Sprintf("invalid payload: %v", err)}, nil
		}
		result, err := h.svc.Buff(ctx, &payload)
		return h.makeResponse(frameType, result, err, true)
	case "attack":
		var payload handler.AttackPayload
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return outboundFrame{Type: frameType, Scope: scopeAck, Error: fmt.Sprintf("invalid payload: %v", err)}, nil
		}
		result, err := h.svc.Attack(ctx, &payload)
		return h.makeResponse(frameType, result, err, true)
	case "trade":
		var payload handler.TradePayload
		if err := json.Unmarshal(frame.Payload, &payload); err != nil {
			return outboundFrame{Type: frameType, Scope: scopeAck, Error: fmt.Sprintf("invalid payload: %v", err)}, nil
		}
		result, err := h.svc.Trade(ctx, &payload)
		return h.makeResponse(frameType, result, err, false)
	default:
		return outboundFrame{Type: frameType, Scope: scopeAck, Error: fmt.Sprintf("unsupported type: %s", frame.Type)}, nil
	}
}

func (h *Handler) makeResponse(frameType string, result interface{}, err error, broadcast bool) (outboundFrame, []outboundFrame) {
	resp := outboundFrame{Type: frameType, Scope: scopeAck}
	if err != nil {
		resp.Error = err.Error()
		return resp, nil
	}
	resp.Result = result
	if !broadcast {
		return resp, nil
	}
	return resp, []outboundFrame{{
		Type:   frameType,
		Scope:  scopeBroadcast,
		Result: result,
	}}
}
