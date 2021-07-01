package broker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	broker "github.com/xn3cr0nx/bitgodine/internal/storage/broker/kafka"
)

func TestKafka(t *testing.T) {
	suite.Run(t, new(TestKafkaSuite))
}

type TestKafkaSuite struct {
	suite.Suite
	kafka *broker.Kafka
}

func (s *TestKafkaSuite) SetupSuite() {
	kafka, _ := broker.NewKafka([]string{"localhost:9092"}, "test")
	s.kafka = kafka

	err := s.kafka.Push(context.Background(), "test", "test")
	s.Nil(err)
}

func (s *TestKafkaSuite) TearDownSuite() {
	s.kafka.Close()
}

func (s *TestKafkaSuite) TestPush() {
	err := s.kafka.Push(context.Background(), "test", "test")
	s.Nil(err)
}

func (s *TestKafkaSuite) TestPull() {
	m, err := s.kafka.Pull(context.Background())
	s.Nil(err)
	s.Equal(string(m.Key), "test")
}
