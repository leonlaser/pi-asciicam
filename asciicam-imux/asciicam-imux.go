package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/websocket"
)

var (
	source   string
	port     int
	channels []chan string
)

func arguments() {
	flag.StringVar(&source, "s", "", "the source websocket e.g. 10.0.0.5:8000")
	flag.IntVar(&port, "p", 8000, "port to accept websocket connections")
	flag.Parse()

	if envSource := os.Getenv("SOURCE"); envSource != "" {
		source = envSource
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		p, _ := strconv.Atoi(envPort)
		port = p
	}
}

func stream(ws *websocket.Conn) {
	fmt.Println("New connection from", ws.Request().RemoteAddr)
	c := make(chan string)
	channels = append(channels, c)
	for ascii := range c {
		websocket.Message.Send(ws, ascii)
	}
	ws.Close()
}

func receive() {

	var message string
	url := fmt.Sprintf("ws://%s/", source)
	ws, err := websocket.Dial(url, "", "http://localhost/")

	if err != nil {
		fmt.Println("Could not connect")
		os.Exit(1)
	}

	for {
		if nil != websocket.Message.Receive(ws, &message) {
			fmt.Println("Connection lost to ", source)
			break
		}
		for _, ch := range channels {
			ch <- message
		}
	}
}

func main() {
	arguments()

	fmt.Println("Settings")
	fmt.Println("  Source:", source)
	fmt.Println("  Port:", port)

	go receive()

	// create a websocket server
	http.Handle("/", websocket.Handler(stream))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
