package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	ping "github.com/t0stbrot/go-ping"
	tracert "github.com/t0stbrot/go-tracert"
	"golang.org/x/sys/windows/svc"
)

type Message struct {
	Type    string `json:"type"`
	Action  string `json:"action,omitempty"`
	Target  string `json:"target,omitempty"`
	Content string `json:"content,omitempty"`
}

type myService struct{}

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}
	initDone := make(chan struct{})
	go func() {
		err := runPrClient()
		if err != nil {
			fmt.Printf("\nWebSocket client error: %v", err)
			s <- svc.Status{State: svc.StopPending}
			close(initDone)
			return
		}
		close(initDone)
	}()
	select {
	case <-initDone:
		s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	case <-time.After(10 * time.Second):
		fmt.Println("Initialization timed out.")
		s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	}
	for {
		select {
		case change := <-r:
			switch change.Cmd {
			case svc.Interrogate:
				s <- change.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				return false, 0
			}
		}
	}
}

func runService(name string, handler svc.Handler) {
	err := svc.Run(name, handler)
	if err != nil {
		fmt.Printf("\nFailed to run service: %v", err)
	}
}

func main() {
	isSvc, _ := svc.IsWindowsService()

	if isSvc {
		runService("WinPR", &myService{})
		return
	}

	fmt.Println("\nRunning in debug mode.")
	err := runPrClient()
	if err != nil {
		fmt.Printf("\nWebSocket client error: %v", err)
	}
}

func traceroute(conn *websocket.Conn, target string) {
	res := tracert.Traceroute4(target, 30, 1000)
	jsonOutput, _ := json.Marshal(res)
	returnMsg := Message{Type: "result", Action: "traceroute", Target: target, Content: string(jsonOutput)}
	conn.WriteJSON(returnMsg)
}

func icmp(conn *websocket.Conn, target string) {
	res := ping.Ping4(target, 64, 5000)
	jsonOutput, _ := json.Marshal(res)
	returnMsg := Message{Type: "result", Action: "icmp", Target: target, Content: string(jsonOutput)}
	conn.WriteJSON(returnMsg)
}

func runPrClient() error {
	// Server to connect to
	serverURL := "wss://winpr.t0stbrot.net/probe"

	for {
		// Dialing the TCP address and trying to connect, then deferring the closing part
		conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
		if err != nil {
			fmt.Printf("\nfailed connect to master; retry in 1 seconds")
			time.Sleep(1 * time.Second)
			continue
		}
		defer conn.Close()

		// Registering at the central service
		registerMsg := Message{Type: "register"}
		conn.WriteJSON(registerMsg)

		fmt.Printf("\nconnected to master")

		// ticker for keepalive
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				returnMsg := Message{Type: "keepalive"}
				conn.WriteJSON(returnMsg)
			}
		}()

		// for loop for reading of the socket
		for {
			// listen
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("\nconnection closed; reconnecting")
				break
			}

			// parse message
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Printf("\nFailed to parse message: %v", err)
				continue
			}

			if msg.Target != "" && msg.Action != "" && msg.Type == "request" {
				if msg.Action == "icmp" {
					fmt.Printf("\nrequested %v for %v", msg.Action, msg.Target)
					icmp(conn, msg.Target)
				} else if msg.Action == "traceroute" {
					fmt.Printf("\nrequested %v for %v", msg.Action, msg.Target)
					traceroute(conn, msg.Target)
				}
				continue
			}
		}
		// if the for loop breaks, wait for 1 seconds for the top-level for loop to restart the connection
		time.Sleep(1 * time.Second)
	}
}
