package messaging

type MessagingClient interface {
    Publish(userID string, messageType string, payload interface{}) error
}

type Message struct {
    UserID  string      `json:"user_id"`
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}
