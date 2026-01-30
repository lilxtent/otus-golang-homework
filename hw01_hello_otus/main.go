package main

import (
	"fmt"

	"golang.org/x/example/hello/reverse"
)

func main() {
	reversedMessage := reverse.String("Hello, OTUS!")
	fmt.Println(reversedMessage)
}
