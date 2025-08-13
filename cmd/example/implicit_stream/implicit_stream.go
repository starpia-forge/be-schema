package main

import (
	"log"

	"github.com/starpia-forge/be-schema"
)

func main() {
	data := []byte(")]}'\r\n\r\n89\r\n[[\"test1\",\"test2\",\"[[null]]\",null,null,null,\"test3\"],[\"test4\",1],[\"test5\",2,\"test6\",3]]\r\n92\r\n[[\"test7\",\"test8\",\"[[null]]\",null,null,null,\"test9\"],[\"test10\",1],[\"test11\",2,\"test12\",3]]\r\n")

	stream, err := beschema.UnmarshalImplicitStream(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Unmarshal: %+v\n", stream)

	data, err = beschema.MarshalImplicitStream(stream)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Marshal: %s\n", string(data))
}
