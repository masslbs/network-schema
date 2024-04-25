package main

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
)

func read(fname string) {
	data, err := os.ReadFile(fname)
	check(err)

	var task Task

	err = proto.Unmarshal(data, &task)
	check(err)

	fmt.Println(task.DoneBy)

	switch task.DoneBy.TypeUrl {
	case "type.googleapis.com/tutorial.Person":
		var r Person
		err = task.DoneBy.UnmarshalTo(&r)
		check(err)
		fmt.Println("age:", r.Age)

	case "type.googleapis.com/tutorial.Robot":
		var r Robot
		err = task.DoneBy.UnmarshalTo(&r)
		check(err)
		fmt.Println("features:", r.Features)

	default:
		fmt.Println("unknown type")
	}

}
