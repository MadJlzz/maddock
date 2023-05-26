package main

import (
	"context"
	"github.com/MadJlzz/maddock/internal/core"
	"log"
	"os"
	"os/signal"
)

func main() {
	cfg, err := core.NewAgentConfiguration("configs/maddock.yml")
	if err != nil {
		log.Fatalf("could not load agent configuration. %v", err)
	}

	agent, err := core.NewAgent(cfg)
	if err != nil {
		log.Fatalf("could not create agent. %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	agent.Start(ctx)
}
