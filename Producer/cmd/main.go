package main

import (
	"fmt"

	"github.com/LootNex/OrderService/Producer/internal/server"
)

func main() {
	if err := server.StartServer(); err != nil {
		fmt.Printf("cannot start Producer server err:%v", err)
	}

}
