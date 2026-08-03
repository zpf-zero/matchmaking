package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"hunjie/internal/pkg/logger"
)

const (
	readTimeout  = 60 * time.Second // 客户端 25s 一次心跳，60s 无消息即断开
	writeTimeout = 10 * time.Second
	maxMsgSize   = 4096
	sendBufSize  = 64
)

// Envelope 收发消息的统一信封
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// OnMessage 业务处理回调，由 service 层在启动时注入。
//
// ws 包只管连接和推送，不碰业务；service 需要推送时调用 ws.Push。
// 用函数变量而不是直接 import，避免 ws ↔ service 循环依赖。
var OnMessage func(userID uint64, msgType string, data json.RawMessage)

// H 全局 Hub。单机内存维护连接，不做多实例广播
var H = &Hub{clients: make(map[uint64]map[*Client]bool)}

type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]map[*Client]bool
}

type Client struct {
	conn   *websocket.Conn
	userID uint64
	send   chan []byte
	once   sync.Once
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 小程序不是浏览器环境，没有同源概念，这里放行
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Serve 升级连接并开始收发
func Serve(c *gin.Context, userID uint64) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("WebSocket 升级失败 uid=%d: %v", userID, err)
		return
	}

	client := &Client{
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, sendBufSize),
	}
	H.add(client)
	logger.Debugf("WebSocket 已连接 uid=%d 在线数=%d", userID, H.OnlineCount())

	go client.writePump()
	go client.readPump()
}

// Push 向指定用户的所有在线连接推送消息
func Push(userID uint64, msgType string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("推送序列化失败: %v", err)
		return
	}
	payload, err := json.Marshal(Envelope{Type: msgType, Data: raw})
	if err != nil {
		return
	}

	H.mu.RLock()
	defer H.mu.RUnlock()
	for c := range H.clients[userID] {
		select {
		case c.send <- payload:
		default:
			// 发送缓冲满说明这个连接卡住了，直接丢弃这条，不阻塞其他人
			logger.Warnf("发送缓冲已满，丢弃消息 uid=%d", userID)
		}
	}
}

// IsOnline 用户是否在线
func IsOnline(userID uint64) bool {
	H.mu.RLock()
	defer H.mu.RUnlock()
	return len(H.clients[userID]) > 0
}

func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.userID] == nil {
		h.clients[c.userID] = make(map[*Client]bool)
	}
	h.clients[c.userID][c] = true
}

func (h *Hub) remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.userID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, c.userID)
		}
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		H.remove(c)
		close(c.send)
		_ = c.conn.Close()
	})
}

func (c *Client) readPump() {
	defer c.close()

	c.conn.SetReadLimit(maxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(readTimeout))

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Debugf("WebSocket 异常断开 uid=%d: %v", c.userID, err)
			}
			return
		}
		// 收到任何消息都算活着，续上超时
		_ = c.conn.SetReadDeadline(time.Now().Add(readTimeout))

		var env Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			c.pushError("消息格式错误")
			continue
		}

		if env.Type == "ping" {
			c.pushRaw(Envelope{Type: "pong"})
			continue
		}
		if OnMessage != nil {
			OnMessage(c.userID, env.Type, env.Data)
		}
	}
}

func (c *Client) writePump() {
	defer c.close()

	for payload := range c.send {
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

func (c *Client) pushRaw(env Envelope) {
	payload, err := json.Marshal(env)
	if err != nil {
		return
	}
	select {
	case c.send <- payload:
	default:
	}
}

func (c *Client) pushError(msg string) {
	raw, _ := json.Marshal(map[string]string{"message": msg})
	c.pushRaw(Envelope{Type: "error", Data: raw})
}
