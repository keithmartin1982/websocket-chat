package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Session struct {
	mutex    sync.Mutex
	MsgCount int64
	ID       string
	Created  time.Time
	Password []byte
	clients  map[string]*websocket.Conn
}

func (s *Session) addClient(id string, c *websocket.Conn) {
	s.mutex.Lock()
	s.clients[id] = c
	s.mutex.Unlock()
	return
}

func (s *Session) removeClient(id string) {
	s.mutex.Lock()
	delete(s.clients, id)
	s.mutex.Unlock()
	return
}

func (s *Session) stats() {
	s.mutex.Lock()
	elapsedTime := time.Now().Sub(s.Created)
	log.Printf("Session %s is dead, removing...\n  Lifetime: %v\n  Message count: %d\n  MPS: %.2f\n", s.ID, elapsedTime, s.MsgCount, float64(s.MsgCount)/elapsedTime.Seconds())
	s.mutex.Unlock()
}

func (s *Session) sendMessage(message []byte, sender string, msgType int) error {
	s.mutex.Lock()
	s.MsgCount++
	for key, conn := range s.clients {
		if key == sender {
			continue
		}
		// TODO : data race
		err := conn.WriteMessage(msgType, message)
		if err != nil {
			return err
		}
	}
	s.mutex.Unlock()
	return nil
}

func (s *Session) connections() (cc int) {
	s.mutex.Lock()
	cc = len(s.clients)
	s.mutex.Unlock()
	return
}

func (s *Session) reportUserCount() {
	var cc int
	cc = s.connections()
	jcc := struct {
		CC int `json:"cc"`
	}{
		CC: cc,
	}
	// TODO : rename var
	tucc, err := json.Marshal(jcc)
	if err != nil {
		log.Printf("Error marshaling jcc: %v\n", err)
	}
	if err := s.sendMessage(tucc, "server", websocket.BinaryMessage); err != nil {
		log.Printf("Error sending message: %v", err)
	}
	
}
