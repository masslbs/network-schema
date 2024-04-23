package main

import "os"

func main() {
	if len(os.Args) != 3 {
		panic("Usage: main (read|write) <file>")
	}
	fname := os.Args[2]
	switch os.Args[1] {
	case "write":
		write(fname)
	case "read":
		read(fname)
	default:
		panic("Usage: main (read|write) <file>")
	}
}
