package main

import (
	"fmt"
	"github.com/MadJlzz/maddock/internal/agent"
)

func main() {
	srv := agent.NewServer()
	if err := srv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
