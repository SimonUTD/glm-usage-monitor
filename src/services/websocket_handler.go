package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID              string
	Connection      *websocket.Conn
	SendChannel     chan WebSocketMessage
	LastPing        time.Time
	IsAuthenticated bool
	UserID          string
}

// WebSocketHandler manages WebSocket connections
type WebSocketHandler struct {
	upgrader            websocket.Upgrader
	clients             map[string]*WebSocketClient
	clientsMutex        sync.RWMutex
	registerChan        chan *WebSocketClient
	unregisterChan      chan *WebSocketClient
	broadcastChan       chan WebSocketMessage
	notificationService *NotificationService
	stopChan            chan bool
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(notificationService *NotificationService) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
		clients:             make(map[string]*WebSocketClient),
		registerChan:        make(chan *WebSocketClient),
		unregisterChan:      make(chan *WebSocketClient),
		broadcastChan:       make(chan WebSocketMessage, 256),
		notificationService: notificationService,
		stopChan:            make(chan bool),
	}
}

// Start starts the WebSocket handler
func (wh *WebSocketHandler) Start() {
	go wh.handleConnections()
	go wh.handleMessages()
	log.Println("WebSocket handler started")
}

// Stop stops the WebSocket handler
func (wh *WebSocketHandler) Stop() {
	close(wh.stopChan)

	wh.clientsMutex.Lock()
	for _, client := range wh.clients {
		client.Connection.Close()
		close(client.SendChannel)
	}
	wh.clientsMutex.Unlock()

	log.Println("WebSocket handler stopped")
}

// HandleWebSocket handles WebSocket connection requests
func (wh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Generate client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	client := &WebSocketClient{
		ID:              clientID,
		Connection:      conn,
		SendChannel:     make(chan WebSocketMessage, 256),
		LastPing:        time.Now(),
		IsAuthenticated: false,
	}

	// Register client
	wh.registerChan <- client

	// Start client goroutines
	go wh.writePump(client)
	go wh.readPump(client)
}

// handleConnections manages client registration and unregistration
func (wh *WebSocketHandler) handleConnections() {
	for {
		select {
		case client := <-wh.registerChan:
			wh.clientsMutex.Lock()
			wh.clients[client.ID] = client
			wh.clientsMutex.Unlock()

			// Register with notification service
			wh.notificationService.RegisterConnection(client.ID, client)

			log.Printf("WebSocket client connected: %s (total: %d)", client.ID, len(wh.clients))

			// Send welcome message
			welcomeMsg := WebSocketMessage{
				Type:      "welcome",
				Data:      map[string]interface{}{"client_id": client.ID},
				Timestamp: time.Now().Unix(),
			}
			select {
			case client.SendChannel <- welcomeMsg:
			default:
				close(client.SendChannel)
			}

		case client := <-wh.unregisterChan:
			wh.clientsMutex.Lock()
			if _, exists := wh.clients[client.ID]; exists {
				delete(wh.clients, client.ID)
				wh.clientsMutex.Unlock()

				// Unregister from notification service
				wh.notificationService.UnregisterConnection(client.ID)

				client.Connection.Close()
				close(client.SendChannel)

				log.Printf("WebSocket client disconnected: %s (total: %d)", client.ID, len(wh.clients))
			} else {
				wh.clientsMutex.Unlock()
			}

		case <-wh.stopChan:
			return
		}
	}
}

// handleMessages handles broadcast messages
func (wh *WebSocketHandler) handleMessages() {
	for {
		select {
		case message := <-wh.broadcastChan:
			wh.clientsMutex.RLock()
			for _, client := range wh.clients {
				select {
				case client.SendChannel <- message:
				default:
					// Client send channel is full, close connection
					wh.unregisterChan <- client
				}
			}
			wh.clientsMutex.RUnlock()

		case <-wh.stopChan:
			return
		}
	}
}

// writePump handles writing messages to WebSocket connection
func (wh *WebSocketHandler) writePump(client *WebSocketClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		wh.unregisterChan <- client
	}()

	for {
		select {
		case message, ok := <-client.SendChannel:
			if !ok {
				client.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Connection.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error for client %s: %v", client.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping
			if err := client.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-wh.stopChan:
			return
		}
	}
}

// readPump handles reading messages from WebSocket connection
func (wh *WebSocketHandler) readPump(client *WebSocketClient) {
	defer func() {
		wh.unregisterChan <- client
	}()

	client.Connection.SetReadLimit(512)
	client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Connection.SetPongHandler(func(string) error {
		client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastPing = time.Now()
		return nil
	})

	for {
		_, messageBytes, err := client.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for client %s: %v", client.ID, err)
			}
			break
		}

		// Parse message
		var message WebSocketMessage
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Invalid WebSocket message from client %s: %v", client.ID, err)
			continue
		}

		// Handle message
		wh.handleClientMessage(client, message)
	}
}

// handleClientMessage handles messages from clients
func (wh *WebSocketHandler) handleClientMessage(client *WebSocketClient, message WebSocketMessage) {
	switch message.Type {
	case "auth":
		// Handle authentication
		if authData, ok := message.Data.(map[string]interface{}); ok {
			if token, ok := authData["token"].(string); ok {
				// Validate token and set user ID
				// This is a placeholder - implement proper token validation
				if token == "valid_token" {
					client.IsAuthenticated = true
					client.UserID = "user_123" // Get from token validation

					response := WebSocketMessage{
						Type:      "auth_success",
						Data:      map[string]interface{}{"user_id": client.UserID},
						Timestamp: time.Now().Unix(),
					}
					select {
					case client.SendChannel <- response:
					default:
					}
				} else {
					response := WebSocketMessage{
						Type:      "auth_error",
						Data:      map[string]interface{}{"error": "Invalid token"},
						Timestamp: time.Now().Unix(),
					}
					select {
					case client.SendChannel <- response:
					default:
					}
				}
			}
		}

	case "ping":
		// Handle ping
		response := WebSocketMessage{
			Type:      "pong",
			Data:      map[string]interface{}{},
			Timestamp: time.Now().Unix(),
		}
		select {
		case client.SendChannel <- response:
		default:
		}

	case "subscribe":
		// Handle subscription to specific events
		if !client.IsAuthenticated {
			wh.sendError(client, "Authentication required")
			return
		}

		// Handle subscription logic
		response := WebSocketMessage{
			Type:      "subscribed",
			Data:      map[string]interface{}{"message": "Subscribed to notifications"},
			Timestamp: time.Now().Unix(),
		}
		select {
		case client.SendChannel <- response:
		default:
		}

	default:
		wh.sendError(client, fmt.Sprintf("Unknown message type: %s", message.Type))
	}
}

// sendError sends an error message to client
func (wh *WebSocketHandler) sendError(client *WebSocketClient, errorMessage string) {
	response := WebSocketMessage{
		Type:      "error",
		Data:      map[string]interface{}{"error": errorMessage},
		Timestamp: time.Now().Unix(),
	}
	select {
	case client.SendChannel <- response:
	default:
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (wh *WebSocketHandler) BroadcastMessage(messageType string, data interface{}) {
	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	select {
	case wh.broadcastChan <- message:
	default:
		log.Printf("Broadcast channel full, message dropped")
	}
}

// SendToClient sends a message to a specific client
func (wh *WebSocketHandler) SendToClient(clientID, messageType string, data interface{}) error {
	wh.clientsMutex.RLock()
	client, exists := wh.clients[clientID]
	wh.clientsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	select {
	case client.SendChannel <- message:
		return nil
	default:
		return fmt.Errorf("client send channel full")
	}
}

// GetClientCount returns the number of connected clients
func (wh *WebSocketHandler) GetClientCount() int {
	wh.clientsMutex.RLock()
	defer wh.clientsMutex.RUnlock()
	return len(wh.clients)
}

// GetConnectedClients returns a list of connected client IDs
func (wh *WebSocketHandler) GetConnectedClients() []string {
	wh.clientsMutex.RLock()
	defer wh.clientsMutex.RUnlock()

	clients := make([]string, 0, len(wh.clients))
	for clientID := range wh.clients {
		clients = append(clients, clientID)
	}

	return clients
}

// GetAuthenticatedClients returns a list of authenticated client IDs
func (wh *WebSocketHandler) GetAuthenticatedClients() []string {
	wh.clientsMutex.RLock()
	defer wh.clientsMutex.RUnlock()

	clients := make([]string, 0)
	for _, client := range wh.clients {
		if client.IsAuthenticated {
			clients = append(clients, client.ID)
		}
	}

	return clients
}

// CleanupInactiveClients removes clients that haven't sent a ping recently
func (wh *WebSocketHandler) CleanupInactiveClients() {
	wh.clientsMutex.Lock()
	defer wh.clientsMutex.Unlock()

	now := time.Now()
	for clientID, client := range wh.clients {
		if now.Sub(client.LastPing) > 5*time.Minute {
			log.Printf("Removing inactive client: %s", clientID)
			client.Connection.Close()
			close(client.SendChannel)
			delete(wh.clients, clientID)
		}
	}
}

// WriteJSON implements WebSocketConnection interface for WebSocketClient
func (wc *WebSocketClient) WriteJSON(v interface{}) error {
	return wc.Connection.WriteJSON(v)
}

// Close implements WebSocketConnection interface for WebSocketClient
func (wc *WebSocketClient) Close() error {
	return wc.Connection.Close()
}
