package main

import (
	"fmt"

	"github.com/LootNex/OrderService/Consumer/internal/server"
)

func main() {
	if err := server.StartServer(); err != nil {
		fmt.Printf("cannot start Consumer server err:%v", err)
	}

}
