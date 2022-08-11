package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		orders, _ := readOrderBookTickerStream()
		fmt.Println(orders)
		time.Sleep(time.Second * 3)
	}
}
