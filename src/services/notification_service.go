package services

import (
	"fmt"
	"log"
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

// NotificationService provides notification management
type NotificationService struct {
	notifications []Notification
	nextID        int
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		notifications: make([]Notification, 0),
		nextID:        1,
	}
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
