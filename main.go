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

func startUp(store map[string]any) {
	storeBytes, err := os.ReadFile("db/komi.json")

	if err != nil {
		log.Fatal("Error reading db ", err)
	}

	err = json.Unmarshal(storeBytes, &store)

	if err != nil {
		log.Fatal("Error encoding store ", err)
	}

	fmt.Println(store)

}

func saveFile(conn net.Conn, key string, value string, store map[string]any) {
	err := os.MkdirAll("db", 0755)
	if err != nil {
		log.Fatal("Error making directory ", err)
	}

	file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_WRONLY, 0755)

	if err != nil {
		log.Fatal("Error creating file ", err)
	}

	defer file.Close()

	for k := range store {
		if k == key {
			conn.Write([]byte("Key already set\n"))
			return
		}
	}
	store[key] = value

	bytes, err := json.Marshal(store)

	if err != nil {
		log.Fatal("Error parsing store ", err)
	}

	_, err = file.Write(bytes)

	if err != nil {
		log.Fatal("Error writing to db ", err)
	}

	conn.Write([]byte("Set Successfully\n"))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Welcome To Komi\n"))
	store := make(map[string]any)

	startUp(store)
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

			if len(parts) < 3 {
				conn.Write([]byte("No value entered\n"))
				continue
			}

			key := parts[1]
			value := strings.Join(parts[2:], " ")

			fmt.Println(key, value)

			saveFile(conn, key, value, store)

		case "LIST":
			for k, v := range store {
				singleDate := fmt.Sprintf("Key: %s\t Value: %s\n", k, v)

				conn.Write([]byte(singleDate))
			}
		case "GET":
			if len(parts) < 2 {
				conn.Write([]byte("No key entered"))
				continue
			}

			key := parts[1]

			found := false
			for k, v := range store {
				if k == key {
					response := fmt.Sprintf("Key: %s\t Value: %s\n", k, v)
					conn.Write([]byte(response))
					found = true
					break
				}

			}
			if !found {
				conn.Write([]byte("No value stored in database for " + key + "\n"))
			}
		case "DEL":

			if len(parts) < 2 {
				conn.Write([]byte("No key entered"))
				continue
			}
			key := parts[1]

			if store[key] == nil {
				conn.Write([]byte("Data not found"))
				continue
			}

			delete(store, key)

			file, err := os.OpenFile("db/komi.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)

			defer file.Close()

			if err != nil {
				log.Fatal("Error opening file", err)
			}

			storeBytes, err := json.Marshal(store)

			if err != nil {
				log.Fatal("Error dencoding store ", err)
			}

			file.Write(storeBytes)

			fmt.Println(store)
			conn.Write([]byte("Deleted Successfully\n"))
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

	for {

		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting connection ", err)
		}

		go handleConnection(conn)
	}
}
