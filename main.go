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

	defer file.Close()

	if err != nil {
		log.Fatal("Error creating file ", err)
	}

	if err != nil {
		log.Fatal("Error reading db ", err)
	}


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

	conn.Write([]byte("Set Successfully\n"))

	if err != nil {
		log.Fatal("Error writing to db ", err)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

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
			storeBytes, err := json.Marshal(store) 
if err != nil {
				log.Fatal("Error reading db ", err)
			}

			conn.Write(storeBytes)
		case "GET":
			if len(parts) < 2 {
				conn.Write([]byte("No key entered"))
				continue
			}

			key := parts[1]

			for k, v := range store {
				if k == key {
					response := fmt.Sprintf("Key: %s\nValue: %s\n", k, v)
					conn.Write([]byte(response))
					return

				}
			}
			conn.Write([]byte("No value stored in database for " + key + "\n"))
		default:
			conn.Write([]byte("Invalid option\n"))
		}

	}
}


func main() {
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

		go handleConnection(conn)
	}
}
