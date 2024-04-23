package main

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
)

func read(fname string) {
	data, err := os.ReadFile(fname)
	check(err)

	var task1 Task

	err = proto.Unmarshal(data, &task1)
	check(err)

	fmt.Println(task1.DoneBy)

	var r Robot
	err = task1.DoneBy.UnmarshalTo(&r)
	check(err)
	fmt.Println(r)
}
