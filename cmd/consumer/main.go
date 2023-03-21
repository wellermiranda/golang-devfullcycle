package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/devfullcycle/gointensivo2/internal/infra/database"
	"github.com/devfullcycle/gointensivo2/internal/usecase"
	"github.com/devfullcycle/gointensivo2/pkg/kafka"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./orders.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repository := database.NewOrderRepository(db)
	usecase := usecase.CalculateFinalPrice{OrderRepository: repository}

	topics := []string{"orders"}
	servers := "host.docker.internal:9094"
	msgChanKafka := make(chan *ckafka.Message)

	go kafka.Consume(topics, servers, msgChanKafka)
	kafkaWorker(msgChanKafka, usecase)
}

func kafkaWorker(msgChan chan *ckafka.Message, uc usecase.CalculateFinalPrice) {
	for msg := range msgChan {
		var OrderInputDTO usecase.OrderInputDTO
		err := json.Unmarshal(msg.Value, &OrderInputDTO)
		if err != nil {
			panic(err)
		}
		outputDto, err := uc.Execute(OrderInputDTO)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Kafka has processed order %s\n", outputDto.ID)
	}
}
