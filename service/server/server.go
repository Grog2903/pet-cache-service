package server

import (
	"bufio"
	"fmt"
	"github.com/Grog2903/pet-cache-service/service/storage"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	store *storage.Storage
}

func NewServer(store *storage.Storage) *Server {
	return &Server{store: store}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка подключения:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}

		parts := strings.Fields(strings.TrimSpace(message))
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SET":
			if len(parts) < 3 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key, value := parts[1], parts[2]
			var ttl time.Duration
			if len(parts) == 4 {
				seconds, _ := strconv.ParseInt(parts[3], 10, 64)
				ttl = time.Duration(seconds) * time.Second
			}

			s.store.Set(key, value, ttl)

			conn.Write([]byte("OK\n"))
		case "GET":
			if len(parts) < 2 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key := parts[1]
			if data, ok := s.store.Get(key); ok {
				conn.Write([]byte(data + "\n"))
			} else {
				conn.Write([]byte("(nil)\n"))
			}
		case "DEL":
			if len(parts) < 2 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key := parts[1]
			s.store.Del(key)

			conn.Write([]byte("OK\n"))
		case "EXISTS":
			if len(parts) < 2 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key := parts[1]
			if s.store.Exists(key) {
				conn.Write([]byte("1\n"))
			} else {
				conn.Write([]byte("0\n"))
			}
		case "EXPIRE":
			if len(parts) < 3 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key := parts[1]
			seconds, _ := strconv.ParseInt(parts[2], 10, 64)
			duration := time.Duration(seconds) * time.Second

			if s.store.Expire(key, duration) {
				conn.Write([]byte("1\n"))
			} else {
				conn.Write([]byte("0\n"))
			}
		case "TTL":
			if len(parts) < 2 {
				conn.Write([]byte("Не хватает аргументов\n"))
				continue
			}

			key := parts[1]
			ttl := s.store.TTL(key)

			conn.Write([]byte(ttl.String() + "\n"))
		default:
			conn.Write([]byte("ERROR: Неизвестная команда\n"))
		}
	}
}
