package sse

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestSSE client.
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
	client, err := NewSource(server.URL, "SSE ID")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 3; i++ {
		event := <-client.Stream
		fmt.Print("SSE<-: ", string(event.Data))
	}
	client.Close()
	time.Sleep(time.Second * 1)
}

// Test Stream
func TestStream(t *testing.T) {
	body := strings.NewReader("id:1 2\nevent:A B\nretry:2\ndata: test\n\ndata: testA\ndata: testB\n\n")
	client := &Client{
		Stream: make(chan Event, 1024),
	}
	go stream(body, client)
	for i := 0; i < 2; i++ {
		event := <-client.Stream
		fmt.Println("ID:", string(event.ID))
		fmt.Println("Type:", string(event.Type))
		fmt.Println("Data:", string(event.Data))
	}
}
