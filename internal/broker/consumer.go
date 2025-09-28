package broker

import (
	"encoding/json"
	"log"
	"wb-order-service/internal/model"
	"wb-order-service/internal/service"

	"github.com/IBM/sarama"
	"github.com/go-playground/validator/v10"
)

type Consumer struct {
	service   *service.OrderService
	validator *validator.Validate
}

func NewConsumer(svc *service.OrderService) *Consumer {
	return &Consumer{
		service:   svc,
		validator: validator.New(),
	}
}

// Start запускает прослушивание топика Kafka в новой горутине
func (c *Consumer) Start(broker, topic string) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer([]string{broker}, config)
	if err != nil {
		log.Fatalf("Ошибка создания Kafka consumer: %v", err)
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Ошибка подписки на партицию Kafka: %v", err)
	}

	log.Printf("Kafka consumer запущен и слушает топик '%s'", topic)

	// Запускаем обработку сообщений в отдельной горутине
	go func() {
		defer consumer.Close()
		defer partitionConsumer.Close()
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				c.handleMessage(msg.Value)
			case err := <-partitionConsumer.Errors():
				log.Printf("Kafka consumer error: %v", err)
			}
		}
	}()
}

// handleMessage обрабатывает одно сообщение из Kafka
func (c *Consumer) handleMessage(data []byte) {
	var order model.Order
	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("ОШИБКА: не удалось декодировать JSON сообщения: %v. Сообщение: %s", err, string(data))
		return // Игнорируем невалидный JSON
	}

	// Валидируем структуру
	if err := c.validator.Struct(order); err != nil {
		log.Printf("ОШИБКА: невалидные данные заказа %s: %v. Сообщение: %s", order.OrderUID, err, string(data))
		return // Игнорируем невалидные данные
	}

	// Если все в порядке, передаем заказ в сервис
	c.service.ProcessNewOrder(order)
}