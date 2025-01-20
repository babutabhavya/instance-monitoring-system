package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type Message struct {
	Port   int    `json:"port"`
	Host   string `json:"host"`
	Status string `json:"status"`
	ID     string `json:"_id"`
}

type Response struct {
	Status        string    `json:"status"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	ID            string    `json:"id"`
}

func publishNATSRequest(payload []byte, natsConn *nats.Conn) {
	err := natsConn.PublishRequest("activity-status.response", "status-queue-group", payload)
	if err != nil {
		log.Printf("Error publishing message %v to NATS: %v", string(payload), err)
	} else {
		log.Printf("Instance published to activity-status.response: %v", string(payload))
	}

}

func getInstanceStatus(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error making request to %s: %v", url, err)
		return "INACTIVE"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "ACTIVE"
	} else if resp.StatusCode == http.StatusInternalServerError {
		return "ERRORED"
	}
	return "INACTIVE"
}

func handleInstanceActivityRequest(msg *nats.Msg) ([]byte, error) {
	fmt.Printf("Received raw message: %s\n", string(msg.Data))
	var message Message
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		log.Printf("Error unmarshaling message: %v\n", err)
		return nil, err
	}
	fmt.Printf("Received message - Host: %s, Port: %v, Status: %s\n", message.Host, message.Port, message.Status)

	url := fmt.Sprintf("http://%s:%v/health", message.Host, message.Port)
	status := getInstanceStatus(url)

	response := Response{
		Status:        status,
		LastUpdatedAt: time.Now(),
		ID:            message.ID,
	}
	fmt.Printf("Instance health response: %s\n", response)

	payload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return nil, err
	}

	return payload, nil
}

func main() {
	url := "nats://nats:4222"
	log.Printf("Connecting to NATS... %s", url)
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	log.Println("Connected to NATS.")
	defer nc.Close()

	sub, err := nc.QueueSubscribe("activity-status.request", "status-queue-group", func(msg *nats.Msg) {
		response, err := handleInstanceActivityRequest(msg)
		if err == nil {
			publishNATSRequest(response, nc)
		}

	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	select {}
}
