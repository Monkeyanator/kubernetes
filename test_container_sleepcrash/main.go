package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Container start...")
	time.Sleep(time.Second * 8)
	panic("Intentional panic")
}
