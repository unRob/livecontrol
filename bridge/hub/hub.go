package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"encoding/json"
	"net"
	"fmt"
	"os"
	"bufio"
	"../message"
)

type Hub struct {
	tcp_server		net.Listener
	master			net.Conn
	connections		map[*websocket.Conn]bool
	Live 			bool

	Broadcast		chan *message.Message
	Control 		chan *message.Message
	Register		chan *websocket.Conn
	Unregister		chan *websocket.Conn
	Ctrl_register	chan net.Conn
	Ctrl_unregister chan bool
}

func NewHub(port string) *Hub {
	addr := fmt.Sprintf(":%s", port)
	tcp_server, err := net.Listen("tcp", addr)

	if err != nil {
		log.Printf("No pude abrir el puerto: %s", err)
		os.Exit(2)
	}

	h := &Hub{
		Live: true,
		tcp_server: tcp_server,
		connections: make(map[*websocket.Conn]bool),
		Broadcast:	make(chan *message.Message, 256),
		Control:	make(chan *message.Message, 256),
		Register:	make(chan *websocket.Conn, 256),
		Unregister:	make(chan *websocket.Conn, 256),
		Ctrl_register: make(chan net.Conn),
		Ctrl_unregister: make(chan bool),
	}

	go h.Run()
	go h.RunTCP()

	return h
}

// func (h *Hub) ListenMaster() {
// 	for {
// 		str, err := h.MasterReader.ReadString('\n')
// 	    if len(str)>0 {
// 	    	msg, err := message.NewMessage([]byte(str))
// 	    	if err != nil {
// 		    	h.Ctrl_unregister <- true
// 	    	}
// 	        h.Broadcast <- msg
// 	    }
// 	    if err!= nil {
// 	        break
// 	    }
// 	}
// }

func (h *Hub) RunTCP() {
	for {
		conn, err := h.tcp_server.Accept()
		if err != nil {
			log.Println("Se cagó el server de TCP")
			log.Println(err)
			os.Exit(2)
		}

		h.Ctrl_register <- conn
		defer conn.Close()
	}
}


func (h *Hub) ReadFromMaster() {
	reader := bufio.NewReader(h.master)
	for {
		if (h.master != nil) {
			str, err := reader.ReadString('\n')

			if err != nil {
				log.Println(err)
				h.Ctrl_unregister <- true
				break
			} else {
				msg, err := message.NewMessage([]byte(str))
				if err != nil {
					log.Println("No pude leer el mensaje")
					log.Println(str)
				} else {
					h.Broadcast <- msg
				}
			}
		} else {
			log.Println("nel")
			break
		}
	}
}


func (h *Hub) Run() {
	for {
		select {
			case ws_conn := <-h.Register:
				// Registro de clientes
				h.connections[ws_conn] = true
				log.Println("Client registered, current count: ", len(h.connections))
			case ws_conn := <-h.Unregister:
				// Channel leave
				if _, ok := h.connections[ws_conn]; ok {
					delete(h.connections, ws_conn)
					// close(c.send)
				}
				log.Println("Client unregistered, current count: ", len(h.connections))
			case conn := <- h.Ctrl_register:
				h.master = conn
				go h.ReadFromMaster()
				h.Live = true
				isLive := message.Message{Evt: "status", Data: "online"}
				go func() {h.Broadcast <- &isLive}()
				log.Println("LIVE: Master registered")
			case <- h.Ctrl_unregister:
				h.master = nil
				h.Live = false
				isntLive := message.Message{Evt: "status", Data: "offline"}
				go func() {h.Broadcast <- &isntLive}()
				log.Println("OFFLINE: Master unregistered")
			case m := <-h.Broadcast:
				log.Println("broadcasting ", m)
				for c := range h.connections {
					go func(c *websocket.Conn) {
						c.WriteJSON(m.AsMap())
					}(c)
				}
			case m := <-h.Control:
				if h.master != nil {
					str, err := json.Marshal(m.AsMap())
					if err != nil {
						log.Println(err)
					} else {
						h.master.Write(str)
					}
				}

		}
	}
}