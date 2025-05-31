package main

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
)

type Data struct {
	Time  time.Time
	Items []string
}

func main() {
	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]any(nil))

	h := &mh

	if os.Args[1] == "enc" {
		f, err := os.Create("test.msgpack")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		enc := codec.NewEncoder(f, h)
		if err := enc.Encode(&Data{
			Time:  time.Now(),
			Items: []string{"hello", "world"},
		}); err != nil {
			panic(err)
		}
	}

	if os.Args[1] == "dec" {
		f, err := os.Open("test.msgpack")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		dec := codec.NewDecoder(f, h)
		var d Data
		if err := dec.Decode(&d); err != nil {
			panic(err)
		}
		fmt.Println("time:", d.Time)
		fmt.Printf("items: %#v\n", d.Items)
	}

}
