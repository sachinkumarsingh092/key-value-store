package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	baseHttpURL = "http://localhost:8080/db/"
	baseWsURL   = "ws://localhost:8080/watch/"
)

func get(baseHttpURL string, key string) string {
	url := baseHttpURL + key

	// make a get request.
	res, err := http.Get(url)
	if err != nil {
		log.Printf("%v: error getting the response\n", err)
	}

	// read the response.
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("%v: error reading the response body\n", err)
	}

	res.Body.Close()

	return string(data)
}

func set(baseHttpURL string, key string, val string) {
	url := baseHttpURL + key

	// request body (payload)
	requestBody := strings.NewReader(val)

	// post some data
	_, err := http.Post(
		url,
		"application/octet-stream",
		requestBody,
	)

	// check for response error
	if err != nil {
		log.Printf("%v: error posting the request\n", err)
	}
}

func watch() {
	c, _, err := websocket.DefaultDialer.Dial(baseWsURL, nil)
	if err != nil {
		log.Print(err)
	}
	defer c.Close()

	// Empty the nolistener channel
	// <-kv.noListener
	// defer func() {
	// 	// When watch returns, unblock calls to another
	// 	// listerers or clients.
	// 	kv.noListener <- struct{}{}
	// }()

	for {
		// receive message
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Print(err)
		}

		log.Printf("Message from websocket: %v\n", string(message))
	}
}

func main() {
	getKey := flag.String("GET", "", "Usage: -GET <key> ")
	setKey := flag.String("SET", "", "Usage: -SET <key> <val>")
	watchflag := flag.Bool("WATCH", false, "Usage: -WATCH")
	flag.Parse()

	setValue := flag.Args()

	switch os.Args[1] {
	case "-GET":
		value := get(baseHttpURL, *getKey)
		fmt.Printf("getKey = %v getval = %v\n", *getKey, value)
	case "-SET":
		set(baseHttpURL, *setKey, setValue[0])
		fmt.Printf("setkey = %v, setvalue = %v\n", *setKey, setValue[0])
	case "-WATCH":
		if *watchflag {
			watch()
		}
	}

}
