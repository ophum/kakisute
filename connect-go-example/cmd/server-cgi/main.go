package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/gin-gonic/gin"
	greetv1 "github.com/ophum/kakisute/connect-go-example/gen/greet/v1"
	"github.com/ophum/kakisute/connect-go-example/gen/greet/v1/greetv1connect"
)

type GreetServer struct{}

func (s *GreetServer) Greet(ctx context.Context, req *connect.Request[greetv1.GreetRequest]) (*connect.Response[greetv1.GreetResponse], error) {
	res := connect.NewResponse(&greetv1.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})
	res.Header().Set("Greet-Version", "v1")
	return res, nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	scriptName := os.Getenv("SCRIPT_NAME")

	greeter := &GreetServer{}
	greetPath, greetHandler := greetv1connect.NewGreetServiceHandler(greeter)

	r := gin.New()
	r.Use(gin.LoggerWithWriter(os.Stderr), gin.Recovery())
	r.Use(stripScriptName(scriptName))

	registerHandler(r, scriptName, greetPath, greetHandler)

	if err := cgi.Serve(r); err != nil {
		log.Fatal(err)
	}
}

func registerHandler(r gin.IRouter, scriptName, path string, handler http.Handler) {
	r.Any(scriptName+path+"*any", gin.WrapH(handler))
}

func stripScriptName(scriptName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request.URL.Path = strings.TrimPrefix(ctx.Request.URL.Path, scriptName)
		log.Println(ctx.Request.URL.Path)
		ctx.Next()
	}
}
