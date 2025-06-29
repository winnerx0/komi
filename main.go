package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	mu    sync.Mutex
	store = make(map[string]any)
)

func startUp(store map[string]any) {

	err := os.MkdirAll("db", 0755)

	if err != nil {
		log.Fatal("Error making directory ", err)
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_RDONLY, 0644)

	if err != nil {
		log.Fatal("Error reading db ", err)
	}

	storeBytes, err := os.ReadFile("db/komi.json")

	if err != nil {
		log.Println("Failed to read file:", err)
		storeBytes = []byte("{}")
	}

	err = json.Unmarshal(storeBytes, &store)

	file.Write(storeBytes)

	if err != nil {
		log.Fatal("Error encoding store ", err)
	}

}

func saveFile(conn net.Conn, key string, value string) {

	file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_WRONLY, 0755)

	if err != nil {
		log.Fatal("Error creating file ", err)
	}

	defer file.Close()

	if _, exists := store[key]; exists {

		conn.Write([]byte("Key already set\n"))
		return

	}

	mu.Lock()

	store[key] = value

	bytes, err := json.Marshal(store)

	if err != nil {
		log.Fatal("Error parsing store ", err)
	}

	mu.Unlock()

	_, err = file.Write(bytes)

	if err != nil {
		log.Fatal("Error writing to db ", err)
	}

	conn.Write([]byte("Set Successfully\n"))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Welcome To Komi\n"))

	// startUp(store)
	opt := make([]byte, 1024)

	for {
		n, err := conn.Read(opt)

		if err == io.EOF {
			fmt.Println("Done reading from server")
		}
		input := string(opt[:n])
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "SET":

			if len(parts) < 3 {
				conn.Write([]byte("No value entered\n"))
				continue
			}

			key := parts[1]

			value := strings.Join(parts[2:], " ")

			fmt.Println(key, value)

			saveFile(conn, key, value)

		case "LIST":
			for k, v := range store {
				singleDate := fmt.Sprintf("Key: %s\t Value: %s\n", k, v)

				conn.Write([]byte(singleDate))
			}
		case "GET":

			if len(parts) < 2 {
				conn.Write([]byte("No key entered\n"))
				continue
			}

			key := parts[1]

			mu.Lock()

			value, ok := store[key]

			mu.Unlock()

			if !ok {
				conn.Write([]byte("No Data Found"))
				continue
			}

			response := fmt.Sprintf("Key: %s\t Value: %s\n", key, value)
			conn.Write([]byte(response))

		case "DEL":

			if len(parts) < 2 {
				conn.Write([]byte("No key entered\n"))
				continue
			}
			key := parts[1]

			if store[key] == nil {
				conn.Write([]byte("Data not found\n"))
				continue
			}

			file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)

			if err != nil {
				log.Fatal("Error opening file", err)
			}

			mu.Lock()

			delete(store, key)

			storeBytes, err := json.Marshal(store)

			if err != nil {
				log.Fatal("Error decoding store ", err)
			}

			mu.Unlock()

			file.Write(storeBytes)

			file.Close()

			conn.Write([]byte("Deleted Successfully\n"))

		case "UPDATE":

			if len(parts) < 3 {
				conn.Write([]byte("No value entered\n"))
				continue
			}
			key := parts[1]

			value := parts[2:]

			mu.Lock()
			_, ok := store[key]

			store[key] = strings.Join(value, " ")
			storeBytes, err := json.Marshal(store)
			mu.Unlock()

			if !ok {
				conn.Write([]byte("Data not found\n"))
				continue
			}

			file, err := os.OpenFile("db/komi.json", os.O_WRONLY, 0755)

			if err != nil {
				log.Fatal("Error opening file", err)
			}

			if err != nil {
				log.Fatal("Error decoding store ", err)
			}

			file.Write(storeBytes)

			file.Close()
			fmt.Println(store)
			conn.Write([]byte("Updating Successfully\n"))

		default:
			conn.Write([]byte("Invalid option\n"))
		}

	}
}

func main() {
	l, err := net.Listen("tcp", ":3005")

	fmt.Println("Listening to server 3005")

	if err != nil {
		log.Fatal("Error starting connection ", err)
	}

	defer l.Close()

	startUp(store)

	for {

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection ", err)
		}

		go handleConnection(conn)
	}
}
