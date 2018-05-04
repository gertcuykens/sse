package sse

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestSSE Conn
func TestSSE(t *testing.T) {
	handler := NewHandler()
	server := httptest.NewTLSServer(handler)
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(time.Second * 1)
			eventString := fmt.Sprintf("%v", time.Now())
			fmt.Println("SSE->:", eventString)
			handler.Push <- []byte("data: " + eventString + "\n\n")
		}
		server.Close()
	}()
	conn, err := NewConn(server.URL, "SSE ID")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 3; i++ {
		event := <-conn.Stream
		fmt.Print("SSE<-: ", string(event.Data))
	}
	conn.Close()
	time.Sleep(time.Second * 1)
}

// Test stream
func TestStream(t *testing.T) {
	body := strings.NewReader("id:1 2\nevent:A B\nretry:2\ndata: test\n\ndata: testA\ndata: testB\n\n")
	conn := &Conn{
		Stream: make(chan Event, 1024),
	}
	go stream(body, conn)
	event := <-conn.Stream
	fmt.Println("ID:", string(event.ID))
	fmt.Println("Type:", string(event.Type))
	fmt.Println("Data:", string(event.Data))
	event = <-conn.Stream
	fmt.Println(string(event.Data))
}
