package message_queue

import (
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/nats-io/nats.go"
)

var (
	natsService *NatsService = nil
)

type NatsService struct {
	conn   *nats.Conn
	Url    string
	logger runtime.Logger
}

func InitNatsService(logger runtime.Logger, natsUrl string) {
	natsService = &NatsService{
		conn:   nil,
		Url:    natsUrl,
		logger: logger,
	}
	natsService.Connect()
}

func GetNatsService() *NatsService {
	return natsService
}

func (conn *NatsService) Connect() {
	var err error
	conn.conn, err = nats.Connect(conn.Url)
	if err != nil {
		conn.logger.Error("Cannot connect to nats server %v", err)
		return
	}
}

func (conn *NatsService) Publish(topic string, data []byte) {
	err := conn.conn.Publish(topic, data)
	if err != nil {
		conn.logger.Error("Publish topic error %v", err)
	}
}

func (conn *NatsService) RegisterAllSubject() {
	for topic, _ := range messageHandler {
		_, err := conn.conn.Subscribe(topic, func(msg *nats.Msg) {
			conn.logger.Info("Receive data %s", topic, string(msg.Data))
			processMessage(msg.Subject, msg.Data)
		})
		if err != nil {
			conn.logger.Error("Subscribe nats topic error %v", err)
		}
	}
}

func (conn *NatsService) Disconnect() {
	conn.conn.Close()
}
