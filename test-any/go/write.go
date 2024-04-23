package main

import (
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func write(fname string) {

	p := Person{
		Name:  "John Doe",
		Id:    1,
		Email: "john@d.oe",
	}

	wrappedP, err := anypb.New(&p)
	check(err)

	r := Robot{
		Name: "R2D2",
		Id:   2,
		Features: []*Robot_Feature{
			{
				Name: "Blue",
			},
		},
	}

	wrappedR, err := anypb.New(&r)
	check(err)

	t1 := Task{
		Title:   "Buy Milk",
		DueDate: timestamppb.Now(),
		DoneBy:  wrappedP,
	}

	t2 := Task{
		Title:   "Buy Bread",
		DueDate: timestamppb.Now(),
		DoneBy:  wrappedR,
	}

	f, err := os.Create(fname)
	check(err)

	t1d, err := proto.Marshal(&t1)
	check(err)

	_, err = f.Write(t1d)
	check(err)

	t2d, err := proto.Marshal(&t2)
	check(err)

	_, err = f.Write(t2d)
	check(err)

	check(f.Close())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
