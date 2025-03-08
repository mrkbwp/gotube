package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"

	"github.com/IBM/sarama"
)

// VideoProcessingMessage представляет сообщение для обработки видео
type VideoProcessingMessage struct {
	VideoID      uuid.UUID `json:"video_id"`
	BucketID     string    `json:"bucket_id"`
	ShardID      string    `json:"shard_id"`
	PathSegment1 string    `json:"path_segment1"`
	PathSegment2 string    `json:"path_segment2"`
	Filename     string    `json:"filename"`
}

// Producer клиент для отправки сообщений в Kafka
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer создает новый Kafka Producer
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Timeout = 10 * time.Second

	fmt.Println(brokers)
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendVideoProcessingMessage отправляет сообщение для обработки видео
func (p *Producer) SendVideoProcessingMessage(ctx context.Context, msg VideoProcessingMessage) error {
	// Сериализуем сообщение в JSON
	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Создаем сообщение для Kafka
	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(msg.VideoID.String()),
		Value: sarama.ByteEncoder(value),
	}

	// Отправляем сообщение в Kafka
	_, _, err = p.producer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	return p.producer.Close()
}
