package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type MovieEvent struct {
	MovieId     int      `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserId      int      `json:"user_id"`
	Rating      float64  `json:"rating"`
	Genres      []string `json:"genres"`
	Description string   `json:"description"`
}

type UserEvent struct {
	UserId    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

type PaymentEvent struct {
	PaymentId  int     `json:"payment_id"`
	UserId     int     `json:"user_id"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	Timestamp  string  `json:"timestamp"`
	MethodType string  `json:"method_type"`
}

type EventResponse struct {
	Status    string `json:"status"`
	Partition int    `json:"partition"`
	Offset    int    `json:"offset"`
	Event     Event  `json:"event"`
}

type Event struct {
	Id        string `json:"id"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Payload   any    `json:"payload"`
}

type Server struct {
	KafkaWriter *kafka.Writer
}

func (s *Server) handleMovie(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var movieEvent MovieEvent
		if err := json.NewDecoder(r.Body).Decode(&movieEvent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		event := Event{
			Id:        fmt.Sprintf("movie-%d-%s", movieEvent.MovieId, movieEvent.Action),
			Type:      "movie",
			Timestamp: time.Now().Format(time.RFC3339),
			Payload:   movieEvent,
		}
		err := s.produceEvent(MovieTopic, event, strconv.Itoa(movieEvent.MovieId))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(EventResponse{
			Status: "success",
			Event:  event,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var userEvent UserEvent
		if err := json.NewDecoder(r.Body).Decode(&userEvent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		event := Event{
			Id:        fmt.Sprintf("user-%d-%s", userEvent.UserId, userEvent.Action),
			Type:      "user",
			Timestamp: userEvent.Timestamp,
			Payload:   userEvent,
		}
		err := s.produceEvent(UserTopic, event, strconv.Itoa(userEvent.UserId))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(EventResponse{
			Status: "success",
			Event:  event,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePayment(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var paymentEvent PaymentEvent
		if err := json.NewDecoder(r.Body).Decode(&paymentEvent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		event := Event{
			Id:        fmt.Sprintf("payment-%d-%s", paymentEvent.PaymentId, paymentEvent.Status),
			Type:      "payment",
			Timestamp: paymentEvent.Timestamp,
			Payload:   paymentEvent,
		}
		err := s.produceEvent(PaymentTopic, event, strconv.Itoa(paymentEvent.UserId))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(EventResponse{
			Status: "success",
			Event:  event,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) produceEvent(topic string, event Event, key string) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = s.KafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: message,
	})
	if err != nil {
		return err
	}

	return nil
}

const (
	MovieTopic   = "movie-events"
	UserTopic    = "user-events"
	PaymentTopic = "payment-events"
)

func main() {
	port := getEnv("PORT", "8082")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:9092")

	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers),
		Balancer: &kafka.LeastBytes{},
	}
	defer kafkaWriter.Close()

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaBrokers},
		GroupID:     "cinemaabyss-events-consumer",
		GroupTopics: []string{MovieTopic, UserTopic, PaymentTopic},
		MaxBytes:    10e6,
	})
	defer kafkaReader.Close()

	go consumeEvents(kafkaReader)

	server := &Server{
		KafkaWriter: kafkaWriter,
	}

	http.HandleFunc("/api/events/movie", server.handleMovie)
	http.HandleFunc("/api/events/user", server.handleUser)
	http.HandleFunc("/api/events/payment", server.handlePayment)
	http.HandleFunc("/api/events/health", handleHealth)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func consumeEvents(kafkaReader *kafka.Reader) {
	for {
		message, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf(
			"Message at topic/partition/offset %v/%v/%v: %s = %s\n",
			message.Topic,
			message.Partition,
			message.Offset,
			string(message.Key),
			string(message.Value),
		)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}
