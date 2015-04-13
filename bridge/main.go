package main

import (
    "log"
    "os"
    "os/signal"
    "net/http"
    "time"
    "io"
    "fmt"
    "unicode/utf8"
    "net"
    "errors"
    "github.com/andrewtj/dnssd"
    "github.com/gorilla/websocket"
    "./message"
    "./hub"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

const VERSION = "1.0"
var h = hub.NewHub(os.Args[1])

func upgradeToWS(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
    if r.Method != "GET" {
        log.Println("Method mismatch")
        http.Error(w, "Method not allowed", 405)
        return &websocket.Conn{}, errors.New("405")
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade failed")
        return &websocket.Conn{}, err
    }

    log.Println("Connection upgraded")
    return conn, nil
}

func parseMessage(conn *websocket.Conn) (*message.Message, error) {
    mt, b, err := conn.ReadMessage()
    var nilmsg = message.Message{}
    if err != nil {
        return &nilmsg, err
    }

    if mt == websocket.TextMessage {
        if !utf8.Valid(b) {
            conn.WriteControl(websocket.CloseMessage,
                websocket.FormatCloseMessage(websocket.CloseInvalidFramePayloadData, ""),
                time.Time{})
            conn.Close()
            return &nilmsg, errors.New("Invalid UTF-8")
        }

        msg, err := message.NewMessage(b)
        if err != nil {
            return &nilmsg, err
        } else {
            return msg, nil
        }
    } else {
        log.Println("Method mismatch")
        return &nilmsg, errors.New("Can't handle your bytes")
    }
}


func handleClient(w http.ResponseWriter, r *http.Request) {
    conn, err := upgradeToWS(w, r);
    if (err != nil) {
        if err != io.EOF {
            h.Unregister <- conn
        }
        log.Println(err)
        return
    }

    defer conn.Close()
    h.Register <- conn

    for {
        msg, err := parseMessage(conn)
        if err != nil {
            err := conn.WriteJSON(map[string]string{"error": err.Error()})
            if err != nil {
                h.Unregister <- conn
                conn.Close()
                return
            }
        } else {
            if h.Live {
                log.Println("CLIENT: %s", msg)
                h.Control <- msg
            } else {
                conn.WriteJSON(map[string]string{"error": "LiveControl Bridge is not running"})
            }
        }
    }
}


func serveHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Not found.", 404)
        return
    }
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", 405)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    data := fmt.Sprintf("LiveControl Server v%s", VERSION)
    io.WriteString(w, "<html><body>"+data+"</body></html>")
}

func main() {

    f, err := os.OpenFile("/Users/rob/ableton.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        // t.Fatalf("error opening file: %v", err)
        os.Exit(5)
    }
    defer f.Close()

    log.SetOutput(f)

    mdns, port, _ := RegisterMDNS()
    // addr := fmt.Sprintf(":%s", os.Args[1])
    // conn, err := net.Listen("tcp", addr)

    // if err != nil {
    //     log.Printf("No pude abrir el puerto: %s", err)
    //     os.Exit(2)
    // }
    go h.Run()

    http.HandleFunc("/", serveHome)
    http.HandleFunc("/control", handleClient)
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)


    ctrlcHandler := make(chan os.Signal, 1)
    signal.Notify(ctrlcHandler, os.Interrupt)
    for sig := range ctrlcHandler {
        if sig == os.Interrupt {
            mdns.Stop();
            time.Sleep(1e9)
            break
        }
    }
}


func RegisterMDNS() (*dnssd.RegisterOp, int, error){
    l,_ := net.Listen("tcp",":0")
    port := 49170//l.Addr().(*net.TCPAddr).Port
    l.Close()

    op, err := dnssd.StartRegisterOp("LiveControl Server", "_livecontrol._tcp", port, RegisteredMDNS)
    if err != nil {
        log.Printf("Failed to register service: %s", err)
        return op, port, err;
    } else {
        log.Printf("Service running on port %d", port)
    }
    return op, port, err;

}


func RegisteredMDNS(op *dnssd.RegisterOp, err error, add bool, name, serviceType, domain string) {
    if err != nil {
        log.Printf("Service registration failed: %s", err)
        return
    }
    if add {
        log.Printf("Service registered as “%s“ in %s", name, domain)
    } else {
        log.Printf("Service “%s” removed from %s", name, domain)
    }
}