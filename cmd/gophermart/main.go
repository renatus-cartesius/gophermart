package main

import (
	"log"

	"github.com/renatus-cartesius/gophermart/pkg/logger"
)

func main() {
	// ctx := context.Background()

	if err := logger.Initialize("DEBUG"); err != nil {
		log.Fatalln(err)
	}

	logger.Log.Debug("testing")
}
