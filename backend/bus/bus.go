package bus

// Simple in-memory bus for notifications. Handlers can publish notifications
// (persisted) to the bus; websocket listener forwards them to connected clients.

type NotificationMessage struct {
	RecipientID int64
	Payload     []byte
}

var NotificationChan chan NotificationMessage

func init() {
	NotificationChan = make(chan NotificationMessage, 256)
}

// PublishNotification attempts to enqueue a payload for a recipient. If the
// channel is full the notification is dropped to avoid blocking request
// handlers.
func PublishNotification(recipientID int64, payload []byte) {
	select {
	case NotificationChan <- NotificationMessage{RecipientID: recipientID, Payload: payload}:
	default:
		// drop if busy
	}
}
