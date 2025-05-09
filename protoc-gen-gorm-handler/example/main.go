package main

import (
	"log"
	"net"
	"net/http"

	"github.com/ophum/kakisute/protoc-gen-gorm-handler/pb/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	handler, err := user.NewUserHandler(db)
	if err != nil {
		panic(err)
	}

	sv := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			log.Println(r.RemoteAddr, r.Method, r.URL.Path)
		}),
	}
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	if err := sv.Serve(l); err != nil {
		panic(err)
	}
}
