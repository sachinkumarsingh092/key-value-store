package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func get(baseURL string, key string) []byte {
	url := baseURL + key

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

	return data
}

func set(baseURL string, key string, val string) {
	url := baseURL + key

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

func main() {
	baseURL := "http://localhost:8080/db/"

	getKey := flag.String("GET", "", "Usage: --GET <key> ")
	setKey := flag.String("SET", "", "Usage: --SET <key> <val>")
	flag.Parse()

	setValue := flag.Args()

	switch os.Args[1] {
	case "--GET":
		value := string(get(baseURL, *getKey))
		fmt.Printf("getKey = %v getval = %v\n", *getKey, value)
	case "--SET":
		set(baseURL, *setKey, setValue[0])
		fmt.Printf("setkey = %v, setvalue = %v\n", *setKey, setValue[0])
	}

}
