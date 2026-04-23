package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	Conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &NATSClient{Conn: nc}, nil
}

func (n *NATSClient) Publish(userID string, messageType string, payload interface{}) error {
	message := Message{
		UserID:  userID,
		Type:    messageType,
		Payload: payload,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	subject := userID
	if subject == "" {
		subject = fmt.Sprintf("admin.%s", messageType)
	}

	return n.Conn.Publish(subject, data)
}
