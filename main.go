package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type MessageRequest struct {
	Message string `json:"message"`
}

func main() {
	cstr := os.Getenv("CONNECTION_STRING")
	qname := os.Getenv("QUEUE_NAME")

	conn, err := amqp.Dial(cstr)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		qname,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare queue")

	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		var req MessageRequest

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(req.Message),
			},
		)

		failOnError(err, "Failed to publish a message")

		c.JSON(http.StatusOK, gin.H{
			"message": "Message posted",
		})
	})
	r.Run()
}
