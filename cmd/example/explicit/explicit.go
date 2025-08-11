package main

import (
	"log"

	"github.com/starpia-forge/be-schema"
)

type Foo struct {
	Bar Bar
}

type Bar struct {
	Field1 string
	Field2 string
}

func main() {
	data := []byte("89\r\n[[\"test1\",\"test2\",\"[[null]]\",null,null,null,\"test3\"],[\"test4\",1],[\"test5\",2,\"test6\",3]]\r\n")

	f, err := beschema.UnmarshalExplicitSchema[Foo](data)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%+v\n", f)
}
