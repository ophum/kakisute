package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	greetv1 "github.com/ophum/kakisute/connect-go-example/gen/greet/v1"
	"github.com/ophum/kakisute/connect-go-example/gen/greet/v1/greetv1connect"
)

type GreetServer struct{}

func (s *GreetServer) Greet(ctx context.Context, req *connect.Request[greetv1.GreetRequest]) (*connect.Response[greetv1.GreetResponse], error) {
	log.Println("Request headers: ", req.Header())
	res := connect.NewResponse(&greetv1.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})
	res.Header().Set("Greet-Version", "v1")
	return res, nil
}

func main() {
	greeter := &GreetServer{}

	path, handler := greetv1connect.NewGreetServiceHandler(greeter)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	http.ListenAndServe(":8080", mux)
}
