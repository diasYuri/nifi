package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/streadway/amqp"
)

// Pedido representa a estrutura de dados do pedido
type Pedido struct {
	ID         string    `bson:"id"`
	Produto    string    `bson:"produto"`
	Quantidade int       `bson:"quantidade"`
	CriadoEm   time.Time `bson:"criado_em"`
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexão com o MongoDB estabelecida com sucesso!")

	// Configurar conexão RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Erro ao conectar ao RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Erro ao abrir o canal RabbitMQ:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"pedidos", // Nome da fila
		false,     // Durable
		false,     // Delete when unused
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		log.Fatal("Erro ao declarar a fila RabbitMQ:", err)
	}

	// Inserir um pedido a cada 5 segundos
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			id, err := inserePedido(client)
			if err == nil {
				_ = publicaEvento(ch, q, id)
			}
		}
	}
}

func publicaEvento(ch *amqp.Channel, q amqp.Queue, id string) error {
	// Publicar o ID do pedido na fila RabbitMQ
	json, _ := json.Marshal(struct{ Id string }{Id: id})
	err := ch.Publish(
		"",     // Exchange
		q.Name, // Routing key
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		})
	if err != nil {
		log.Println("Erro ao publicar na fila RabbitMQ:", err)
		return err
	}
	fmt.Println("ID do pedido publicado na fila RabbitMQ:", id)
	return nil
}

func inserePedido(client *mongo.Client) (string, error) {
	collection := client.Database("test").Collection("pedidos")
	pedido := Pedido{
		ID:         uuid.New().String(),
		Produto:    fmt.Sprintf("Produto %d", rand.Intn(10)+1),
		Quantidade: rand.Intn(50) + 1,
		CriadoEm:   time.Now(),
	}
	_, err := collection.InsertOne(context.Background(), pedido)
	if err != nil {
		log.Println("Erro ao inserir pedido:", err)
		return "", err
	}
	fmt.Println("Pedido inserido com sucesso!")
	return pedido.ID, nil
}
