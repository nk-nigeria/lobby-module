package message_queue

var (
	messageHandler map[string]func(data []byte)
)

func init() {
	messageHandler = make(map[string]func(data []byte))
}

func RegisterHandler(subject string, handler func(data []byte)) {
	messageHandler[subject] = handler
}

func processMessage(subject string, data []byte) {
	if handler, ok := messageHandler[subject]; ok {
		handler(data)
	}
}
