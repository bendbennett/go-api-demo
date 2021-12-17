package consume

import (
	"net"
	"strconv"

	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/segmentio/kafka-go"
)

func CreateTopics(topicConfigs config.TopicConfigs) error {
	var (
		conn *kafka.Conn
		err  error
	)

	for _, v := range topicConfigs.Brokers {
		conn, err = kafka.Dial("tcp", v)
		if err != nil {
			return err
		}
		if conn != nil {
			break
		}
	}

	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	var controllerConn *kafka.Conn

	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(topicConfigs.Conf...)
	if err != nil {
		return err
	}

	return nil
}
