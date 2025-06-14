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

func saveFile(key string, value string, store map[string]string) {

	err := os.MkdirAll("db", 0755)

	if err != nil {
		log.Fatal("Error making directory ", err)
	}

	file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_WRONLY, 0755)

	defer file.Close()

	if err != nil {
		log.Fatal("Error creating file ", err)
	}

	// store := Store{
	// Key:   key,
	// Value: value,
	// }

	storesBytes, err := os.ReadFile("db/komi.json")

	if err != nil {
		log.Fatal("Error reading db ", err)
	}

	if len(storesBytes) == 0 {

		store = make(map[string]string)
	}

	// for k, _ := range *store {
	// 	if k == key {
	// 		fmt.Println("Key already set")
	// 		os.Exit(1)
	// 	}
	// }
	store[key] = value

	bytes, err := json.Marshal(store)

	if err != nil {
		log.Fatal("Error parsing store ", err)
	}

	_, err = file.Write(bytes)

	if err != nil {
		log.Fatal("Error writing to db ", err)
	}

}

func handleConnection(conn net.Conn, ch chan<- string, store map[string]string) {

	defer conn.Close()

	opt := make([]byte, 1024)

	for {
		n, err := conn.Read(opt)

		if err != nil {
			log.Fatal("Error reading from server ", err)
		}
		input := string(opt[:n])
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "SET":
			key := parts[1]
			value := strings.Join(parts[2:], " ")

			fmt.Println(key)
			ch <- key

			saveFile(key, value, store)
		default:
			fmt.Println("Invalid option")
		}

		fmt.Println(string(opt))
	}

}

func main() {

	ch := make(chan string)

	store := make(map[string]string)

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

		go handleConnection(conn, ch, store)
		fmt.Println(<-ch)
	}

}
