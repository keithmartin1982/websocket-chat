package main

import (
	_ "embed"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	//go:embed root.html
	rootHtml []byte
	upgrader = websocket.Upgrader{}
	// TODO : set domain here \/
	allowedOrigins = "*"
)

type Server struct {
	mutex    sync.Mutex
	port     string
	cert     string
	key      string
	tlsPort  string
	Sessions map[string]*Session
}

func (s *Server) addSession(session *Session) {
	s.mutex.Lock()
	s.Sessions[session.ID] = session
	s.mutex.Unlock()
}

func (s *Server) removeSession(session *Session) {
	s.mutex.Lock()
	delete(s.Sessions, session.ID)
	s.mutex.Unlock()
}

func (s *Server) getSession(id string) (*Session, bool) {
	s.mutex.Lock()
	session, ok := s.Sessions[id]
	s.mutex.Unlock()
	return session, ok
}

func (s *Server) upgrader() func(r *http.Request) bool {
	return func(r *http.Request) bool {
		return true
	}
}

func (s *Server) start() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write(rootHtml); err != nil {
			log.Printf("error writing root html: %v\n", err)
			return
		}
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = s.upgrader()
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer func(c *websocket.Conn) {
			err := c.Close()
			if err != nil {
				log.Printf("Error closing connection: %v", err)
			}
		}(c)
		pmt, pm, err := c.ReadMessage()
		if err != nil {
			log.Printf("Error reading login message: %v", err)
			return
		}
		prs := struct {
			ID string `json:"id"`
			P  string `json:"p"`
		}{}
		if pmt == websocket.TextMessage {
			if err := json.Unmarshal(pm, &prs); err != nil {
				log.Printf("Error parsing login message: %v", err)
				return
			}
		} else {
			return
		}
		if len(prs.P) > 32 || len(prs.P) < 8 || len(prs.ID) > 32 || len(prs.ID) < 3 {
			return
		}
		var ns *Session
		ns, ok := s.getSession(prs.ID)
		if ok {
			// LIMIT connected users to 10
			if ns.connections() >= 10 {
				// TODO : send reason
				return
			}
			if checkPass([]byte(prs.P), ns.Password) != nil {
				return
			}
			ns.addClient(r.RemoteAddr, c)
		} else {
			hash, err := genHash([]byte(prs.P))
			if err != nil {
				log.Printf("Error generating hash: %v\n", err)
				return
			}
			ns = &Session{
				Created:  time.Now(),
				MsgCount: 0,
				ID:       prs.ID,
				clients:  make(map[string]*websocket.Conn),
				Password: hash,
			}
			ns.addClient(r.RemoteAddr, c)
			s.addSession(ns)
		}
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				ns.removeClient(r.RemoteAddr)
				var cc int
				cc = ns.connections()
				if cc == 0 {
					ns.stats()
					s.removeSession(ns)
				} else {
					ns.reportUserCount()
				}
				return
			}
			ns.reportUserCount()
			if mt == websocket.TextMessage {
				if err := ns.sendMessage(message, r.RemoteAddr, websocket.TextMessage); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			} // else {fmt.Printf("Message received: %v\n", string(message))}
		}
	})
	if len(s.cert) < 2 || len(s.key) < 2 {
		log.Printf("Serving: http://127.0.0.1%s/", s.port)
		if err := http.ListenAndServe(s.port, nil); err != nil {
			log.Printf("ListenAndServe: %v\n", err)
		}
	} else {
		go func() {
			if err := http.ListenAndServe(s.port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			})); err != nil {
				log.Printf("Error listen::TLS redirect: %v\n", err)
			}
		}()
		if err := http.ListenAndServeTLS(s.tlsPort, s.cert, s.key, nil); err != nil {
			log.Printf("ListenAndServeTLS: %v\n", err)
		}
	}
}
