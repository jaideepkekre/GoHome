package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var q amqp.Queue
var ch *amqp.Channel

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Printf("%T", sample)
	conn, _ = amqp.Dial("amqp://guest:guest@localhost:5672/")
	ch, _ = conn.Channel()
	q, _ = ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	router := mux.NewRouter()
	fmt.Printf("%T", conn)
	router.HandleFunc("/sample", sample).Methods("POST")
	log.Fatal(http.ListenAndServe(":8001", router))
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func sample(w http.ResponseWriter, r *http.Request) {

	var dummy interface{}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &dummy)
	json_map := dummy.(map[string]interface{})
	byte_payload, _ := json.Marshal(json_map)
	go add2q(byte_payload)

}

func add2q(payload []byte) {
	_ = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        payload,
		})

}
