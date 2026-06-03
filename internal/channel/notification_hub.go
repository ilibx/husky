package channel

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

type NotificationHub struct {
	mu      sync.RWMutex
	clients map[uint]map[chan int64]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{clients: make(map[uint]map[chan int64]struct{})}
}

func (h *NotificationHub) Subscribe(userID uint) chan int64 {
	ch := make(chan int64, 1)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan int64]struct{})
	}
	h.clients[userID][ch] = struct{}{}
	return ch
}

func (h *NotificationHub) Unsubscribe(userID uint, ch chan int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.clients[userID]; clients != nil {
		delete(clients, ch)
		close(ch)
		if len(clients) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *NotificationHub) Publish(userID uint, count int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[userID] {
		select {
		case ch <- count:
		default:
		}
	}
}

func (h *Handler) WatchNotifications(c *gin.Context) {
	userID, ok := h.getUserIDForNotifications(c)
	if !ok {
		return
	}
	if h.notificationHub == nil {
		httputil.Error(c, http.StatusServiceUnavailable, errors.ErrInternal, "notification stream unavailable")
		return
	}

	count, err := h.querySvc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	conn, err := notificationUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := h.notificationHub.Subscribe(userID)
	defer h.notificationHub.Unsubscribe(userID, ch)

	if err := writeUnreadCount(conn, count); err != nil {
		return
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.NextReader(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-done:
			return
		case count := <-ch:
			if err := writeUnreadCount(conn, count); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteJSON(gin.H{"type": "ping", "time": time.Now().Unix()}); err != nil {
				return
			}
		}
	}
}

var notificationUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func writeUnreadCount(conn *websocket.Conn, count int64) error {
	return conn.WriteJSON(gin.H{"type": "unread_count", "count": count})
}

func (h *Handler) getUserIDForNotifications(c *gin.Context) (uint, bool) {
	if userID, exists := c.Get("user_id"); exists {
		uid, ok := userID.(uint)
		if ok {
			return notificationScopeUserID(c, uid), true
		}
	}

	tokenString := c.Query("token")
	if tokenString == "" {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "missing token")
		return 0, false
	}

	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, auth.JWTKeyFunc)
	if err != nil || !token.Valid {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "invalid token")
		return 0, false
	}
	c.Set("role", claims.Role)
	return notificationScopeUserID(c, claims.UserID), true
}

func notificationScopeUserID(c *gin.Context, uid uint) uint {
	role, _ := c.Get("role")
	if role == "admin" && c.Query("scope") != "mine" {
		return 0
	}
	return uid
}
