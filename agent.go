package main

import (
	"context"
	"github.com/MadJlzz/maddock/internal/service"
	"log"
	"os"
	"os/signal"
)

func main() {
	cfg, err := service.NewAgentConfiguration("configs/maddock.yml")
	if err != nil {
		log.Fatalf("could not load agent configuration. %v", err)
	}

	agent, err := service.New(cfg)
	if err != nil {
		log.Fatalf("could not create agent. %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	agent.Start(ctx)
}
