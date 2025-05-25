package main

import (
	"log"

	"github.com/casbin/casbin/v2"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	e, err := casbin.NewEnforcer("./model.conf", "./policy.csv")
	if err != nil {
		return err
	}

	sub := "doe"
	obj := "getHoge"
	act := "read"

	ok, err := e.Enforce(sub, obj, act)
	if err != nil {
		return err
	}

	if ok {
		log.Println("permit")
	} else {
		log.Println("deny")
	}
	return nil
}
