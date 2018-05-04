package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gertcuykens/sse"
)

func main() {
	handler := sse.NewHandler()
	http.Handle("/sse", handler)
	http.Handle("/", http.FileServer(http.Dir("data")))
	go func() {
		for {
			time.Sleep(time.Second * 1)
			eventString := fmt.Sprintf("%v", time.Now())
			fmt.Println("SSE->:", eventString)
			handler.Push <- []byte("data: " + eventString + "\n\n")
		}
	}()
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
