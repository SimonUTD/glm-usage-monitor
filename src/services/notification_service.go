package services

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// NotificationType represents different types of notifications
type NotificationType int

const (
	NotificationTypeInfo NotificationType = iota
	NotificationTypeSuccess
	NotificationTypeWarning
	NotificationTypeError
)

// Notification represents a user notification
type Notification struct {
	ID        int              `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Data      interface{}      `json:"data,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection interface {
	WriteJSON(v interface{}) error
	Close() error
}

// NotificationService provides notification management
type NotificationService struct {
	notifications    []Notification
	nextID           int
	connections      map[string]WebSocketConnection
	connectionsMutex sync.RWMutex
	broadcastChannel chan Notification
	stopBroadcast    chan bool
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	ns := &NotificationService{
		notifications:    make([]Notification, 0),
		nextID:           1,
		connections:      make(map[string]WebSocketConnection),
		broadcastChannel: make(chan Notification, 100),
		stopBroadcast:    make(chan bool),
	}

	// Start broadcast goroutine
	go ns.broadcastLoop()

	return ns
}

// AddNotification adds a new notification
func (ns *NotificationService) AddNotification(notificationType NotificationType, title, message string, data interface{}) {
	notification := Notification{
		ID:        ns.nextID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Data:      data,
		CreatedAt: time.Now(),
	}

	ns.notifications = append(ns.notifications, notification)
	ns.nextID++

	log.Printf("Notification added: %s - %s", title, message)

	// Broadcast to WebSocket connections
	select {
	case ns.broadcastChannel <- notification:
	default:
		log.Printf("Broadcast channel full, notification not sent to WebSocket clients")
	}
}

// AddSyncSuccessNotification adds a sync success notification
func (ns *NotificationService) AddSyncSuccessNotification(billingMonth string, syncedCount, totalCount int) {
	title := "同步完成"
	message := fmt.Sprintf("账单月份 %s 同步完成，成功同步 %d/%d 条记录", billingMonth, syncedCount, totalCount)

	data := map[string]interface{}{
		"billing_month": billingMonth,
		"synced_count":  syncedCount,
		"total_count":   totalCount,
		"type":          "sync_success",
	}

	ns.AddNotification(NotificationTypeSuccess, title, message, data)
}

// AddSyncFailureNotification adds a sync failure notification
func (ns *NotificationService) AddSyncFailureNotification(billingMonth string, errorMessage string) {
	title := "同步失败"
	message := fmt.Sprintf("账单月份 %s 同步失败：%s", billingMonth, errorMessage)

	data := map[string]interface{}{
		"billing_month": billingMonth,
		"error_message": errorMessage,
		"type":          "sync_failure",
	}

	ns.AddNotification(NotificationTypeError, title, message, data)
}

// AddTokenExpiredNotification adds a token expired notification
func (ns *NotificationService) AddTokenExpiredNotification(tokenName string) {
	title := "令牌过期"
	message := fmt.Sprintf("API令牌 %s 已过期，请更新令牌", tokenName)

	data := map[string]interface{}{
		"token_name": tokenName,
		"type":       "token_expired",
	}

	ns.AddNotification(NotificationTypeWarning, title, message, data)
}

// GetUnreadNotifications retrieves all unread notifications
func (ns *NotificationService) GetUnreadNotifications() []Notification {
	var unread []Notification
	for _, notification := range ns.notifications {
		if notification.ReadAt == nil {
			unread = append(unread, notification)
		}
	}
	return unread
}

// MarkAsRead marks notifications as read
func (ns *NotificationService) MarkAsRead(notificationID int) {
	for i, notification := range ns.notifications {
		if notification.ID == notificationID {
			ns.notifications[i].ReadAt = &[]time.Time{time.Now()}[0]
			log.Printf("Notification %d marked as read", notificationID)
			break
		}
	}
}

// MarkAllAsRead marks all notifications as read
func (ns *NotificationService) MarkAllAsRead() {
	now := time.Now()
	for i := range ns.notifications {
		if ns.notifications[i].ReadAt == nil {
			ns.notifications[i].ReadAt = &now
		}
	}
	log.Println("All notifications marked as read")
}

// ClearNotifications clears all notifications
func (ns *NotificationService) ClearNotifications() {
	ns.notifications = make([]Notification, 0)
	ns.nextID = 1
	log.Println("All notifications cleared")
}

// GetNotificationCount returns the count of unread notifications
func (ns *NotificationService) GetNotificationCount() int {
	count := 0
	for _, notification := range ns.notifications {
		if notification.ReadAt == nil {
			count++
		}
	}
	return count
}

// GetRecentNotifications returns recent notifications with limit
func (ns *NotificationService) GetRecentNotifications(limit int) []Notification {
	if limit <= 0 || limit > len(ns.notifications) {
		limit = len(ns.notifications)
	}

	// Return the most recent notifications (in reverse order)
	recent := make([]Notification, limit)
	total := len(ns.notifications)

	for i := 0; i < limit; i++ {
		index := total - 1 - i
		if index >= 0 {
			recent[i] = ns.notifications[index]
		}
	}

	return recent
}

// RegisterConnection registers a WebSocket connection
func (ns *NotificationService) RegisterConnection(clientID string, conn WebSocketConnection) {
	ns.connectionsMutex.Lock()
	defer ns.connectionsMutex.Unlock()

	ns.connections[clientID] = conn
	log.Printf("WebSocket connection registered: %s", clientID)

	// Send unread notifications to new connection
	unreadNotifications := ns.GetUnreadNotifications()
	for _, notification := range unreadNotifications {
		if err := conn.WriteJSON(notification); err != nil {
			log.Printf("Failed to send notification to client %s: %v", clientID, err)
			delete(ns.connections, clientID)
			return
		}
	}
}

// UnregisterConnection removes a WebSocket connection
func (ns *NotificationService) UnregisterConnection(clientID string) {
	ns.connectionsMutex.Lock()
	defer ns.connectionsMutex.Unlock()

	if conn, exists := ns.connections[clientID]; exists {
		conn.Close()
		delete(ns.connections, clientID)
		log.Printf("WebSocket connection unregistered: %s", clientID)
	}
}

// GetConnectionCount returns the number of active WebSocket connections
func (ns *NotificationService) GetConnectionCount() int {
	ns.connectionsMutex.RLock()
	defer ns.connectionsMutex.RUnlock()

	return len(ns.connections)
}

// broadcastLoop broadcasts notifications to all connected WebSocket clients
func (ns *NotificationService) broadcastLoop() {
	for {
		select {
		case notification := <-ns.broadcastChannel:
			ns.broadcastNotification(notification)
		case <-ns.stopBroadcast:
			return
		}
	}
}

// broadcastNotification sends a notification to all connected clients
func (ns *NotificationService) broadcastNotification(notification Notification) {
	ns.connectionsMutex.RLock()
	connections := make(map[string]WebSocketConnection)
	for id, conn := range ns.connections {
		connections[id] = conn
	}
	ns.connectionsMutex.RUnlock()

	for clientID, conn := range connections {
		if err := conn.WriteJSON(notification); err != nil {
			log.Printf("Failed to broadcast notification to client %s: %v", clientID, err)
			ns.UnregisterConnection(clientID)
		}
	}
}

// SendCustomMessage sends a custom message to a specific client
func (ns *NotificationService) SendCustomMessage(clientID string, message interface{}) error {
	ns.connectionsMutex.RLock()
	conn, exists := ns.connections[clientID]
	ns.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not connected", clientID)
	}

	return conn.WriteJSON(message)
}

// BroadcastCustomMessage broadcasts a custom message to all connected clients
func (ns *NotificationService) BroadcastCustomMessage(message interface{}) {
	ns.connectionsMutex.RLock()
	connections := make(map[string]WebSocketConnection)
	for id, conn := range ns.connections {
		connections[id] = conn
	}
	ns.connectionsMutex.RUnlock()

	for clientID, conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Failed to broadcast custom message to client %s: %v", clientID, err)
			ns.UnregisterConnection(clientID)
		}
	}
}

// GetConnectedClients returns a list of connected client IDs
func (ns *NotificationService) GetConnectedClients() []string {
	ns.connectionsMutex.RLock()
	defer ns.connectionsMutex.RUnlock()

	clients := make([]string, 0, len(ns.connections))
	for clientID := range ns.connections {
		clients = append(clients, clientID)
	}

	return clients
}

// Shutdown gracefully shuts down the notification service
func (ns *NotificationService) Shutdown() {
	close(ns.stopBroadcast)

	ns.connectionsMutex.Lock()
	defer ns.connectionsMutex.Unlock()

	for clientID, conn := range ns.connections {
		conn.Close()
		delete(ns.connections, clientID)
		log.Printf("WebSocket connection closed during shutdown: %s", clientID)
	}

	log.Println("Notification service shutdown complete")
}

// AddSyncProgressNotification adds a sync progress notification
func (ns *NotificationService) AddSyncProgressNotification(billingMonth string, current, total int, percentage float64) {
	title := "同步进度"
	message := fmt.Sprintf("账单月份 %s 同步进度：%d/%d (%.1f%%)", billingMonth, current, total, percentage)

	data := map[string]interface{}{
		"billing_month": billingMonth,
		"current":       current,
		"total":         total,
		"percentage":    percentage,
		"type":          "sync_progress",
	}

	ns.AddNotification(NotificationTypeInfo, title, message, data)
}

// AddSystemNotification adds a system-level notification
func (ns *NotificationService) AddSystemNotification(title, message string, notificationType NotificationType) {
	data := map[string]interface{}{
		"type": "system_notification",
	}

	ns.AddNotification(notificationType, title, message, data)
}

// ExportNotifications exports notifications as JSON
func (ns *NotificationService) ExportNotifications() ([]byte, error) {
	return json.MarshalIndent(ns.notifications, "", "  ")
}

// GetNotificationsByType returns notifications filtered by type
func (ns *NotificationService) GetNotificationsByType(notificationType NotificationType) []Notification {
	var filtered []Notification
	for _, notification := range ns.notifications {
		if notification.Type == notificationType {
			filtered = append(filtered, notification)
		}
	}
	return filtered
}

// GetNotificationsByDateRange returns notifications within a date range
func (ns *NotificationService) GetNotificationsByDateRange(start, end time.Time) []Notification {
	var filtered []Notification
	for _, notification := range ns.notifications {
		if notification.CreatedAt.After(start) && notification.CreatedAt.Before(end) {
			filtered = append(filtered, notification)
		}
	}
	return filtered
}
