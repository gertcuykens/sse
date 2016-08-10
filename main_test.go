package sse

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain Handler.
func TestMain(m *testing.M) {
	handler := NewHandler()
	go func() {
		for {
			time.Sleep(time.Second * 1)
			eventString := fmt.Sprintf("%v", time.Now())
			fmt.Println("SSE->:", eventString)
			handler.Push <- []byte("data: " + eventString + "\n\n")
		}
	}()
	go func() {
		log.Fatal("HTTPS server error: ",
			http.ListenAndServeTLS(":8080",
				"public.pem",
				"private.pem",
				handler))
	}()
	os.Exit(m.Run())
}

// TestSSE client.
func TestSSE(t *testing.T) {
	client, err := NewSource("https://localhost:8080/sse", "SSE ID")
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 3; i++ {
		event := <-client.Stream
		fmt.Print("SSE<-: ", string(event.Data))
	}
	client.Close()
	time.Sleep(1 * time.Second)
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
		fmt.Println("Retry: ", client.Err, client.Retry)
	}
}
