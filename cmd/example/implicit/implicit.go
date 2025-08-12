package main

import (
	"log"

	"github.com/starpia-forge/be-schema"
)

func main() {
	data := []byte("89\r\n[[\"test1\",\"test2\",\"[[null]]\",null,null,null,\"test3\"],[\"test4\",1],[\"test5\",2,\"test6\",3]]\r\n")

	entity, err := beschema.UnmarshalImplicitSchema(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v\n", entity)
}
