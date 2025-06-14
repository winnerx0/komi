package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Store struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func saveFile(key string, value string) {
	err := os.MkdirAll("db", 0755)

	if err != nil {
		log.Fatal("Error making directory ", err)
	}

	file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_WRONLY, 0755)

	defer file.Close()

	if err != nil {
		log.Fatal("Error creating file ", err)
	}

	store := Store{
		Key:   key,
		Value: value,
	}

	var stores []Store

	storesBytes, err := os.ReadFile("db/komi.json")

	if err != nil {
		log.Fatal("Error reading db ", err)
	}

	if len(storesBytes) > 0 {

		err = json.Unmarshal(storesBytes, &stores)

		if err != nil {
			log.Fatal("Error parsing stores ", err)
		}

	} else {
		stores = []Store{}
	}
	stores = append(stores, store)
	
	f, err := json.Marshal(&stores)

	if err != nil {
		log.Fatal("Error parsing data ", err)
	}

	_, err = file.Write(f)

	if err != nil {
		log.Fatal("Error writing to db ", err)
	}

}

func handleConnection(conn net.Conn, store chan<- string) {

	defer conn.Close()

	opt := make([]byte, 1024)

	for {
		n, err := conn.Read(opt)

		input := string(opt[:n])
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "SET":
			key := parts[1]
			value := parts[2]

			fmt.Println(key)
			store <- key

			saveFile(key, value)
		default:
			fmt.Println("Invalid option")
		}

		if err != nil {
			log.Fatal("Error reading from server ", err)
		}
		fmt.Println(string(opt))
	}

}

func main() {

	store := make(chan string)

	l, err := net.Listen("tcp", ":3005")

	defer l.Close()

	fmt.Println("Listening to server 3005")

	if err != nil {
		log.Fatal("Error starting connection ", err)
	}

	for {

		conn, err := l.Accept()

		if err != nil {
			log.Fatal("Error accepting connection ", err)
		}

		go handleConnection(conn, store)
		fmt.Println(<-store)
	}

}
