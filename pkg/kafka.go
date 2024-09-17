package pkg

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// KafkaProducer wraps the Sarama SyncProducer
type KafkaProducer struct {
	Producer sarama.SyncProducer
}

// NewKafkaProducer initializes the Kafka producer
func NewKafkaProducer(brokers []string) *KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama producer: %v", err)
	}

	return &KafkaProducer{Producer: producer}
}

// EnqueueTask sends a task message to the specified Kafka topic
func (kp *KafkaProducer) EnqueueTask(task Task, topic string) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(task.ID),
		Value: sarama.ByteEncoder(taskBytes),
	}

	_, _, err = kp.Producer.SendMessage(msg)
	return err
}

// KafkaConsumer wraps the Sarama Consumer
type KafkaConsumer struct {
	Consumer sarama.Consumer
}

// NewKafkaConsumer initializes the Kafka consumer
func NewKafkaConsumer(brokers []string) *KafkaConsumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Sarama consumer: %v", err)
	}

	return &KafkaConsumer{Consumer: consumer}
}

// ConsumeTasks starts consuming tasks from the specified topic
func (kc *KafkaConsumer) ConsumeTasks(topic string, handler func(Task)) {
	partitionList, err := kc.Consumer.Partitions(topic)
	if err != nil {
		log.Fatalf("Failed to get partitions for topic %s: %v", topic, err)
	}

	for _, partition := range partitionList {
		pc, err := kc.Consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Failed to start consuming partition %d: %v", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for message := range pc.Messages() {
				var task Task
				if err := json.Unmarshal(message.Value, &task); err != nil {
					log.Printf("Failed to unmarshal task: %v", err)
					continue
				}
				handler(task)
			}
		}(pc)
	}
}
