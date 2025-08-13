package main

import (
	"log"

	"github.com/starpia-forge/be-schema"
)

type Entity struct {
	Sub1 SubEntity1 `beschema:"1"`
	Sub2 SubEntity2 `beschema:"3"`
}

type SubEntity1 struct {
	Field1 string `beschema:"1"`
	Field2 string `beschema:"2"`
}

type SubEntity2 struct {
	Field1 string `beschema:"3"`
	Field2 string `beschema:"4"`
}

func main() {
	data := []byte("89\r\n[[\"test1\",\"test2\",\"[[null]]\",null,null,null,\"test3\"],[\"test4\",1],[\"test5\",2,\"test6\",3]]\r\n")

	entity, err := beschema.UnmarshalExplicitSchema[Entity](data, true)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Unmarshal: %+v\n", entity)

	data, err = beschema.MarshalExplicitSchema(entity)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Marshal: %+v\n", string(data))
}
