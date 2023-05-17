package core

import (
	"context"
	"github.com/MadJlzz/maddock/internal/modules"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Agent struct {
	cfg     *AgentConfiguration
	pollWg  sync.WaitGroup
	gitPath string
}

func NewAgent(cfg *AgentConfiguration) (*Agent, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return &Agent{
		cfg:     cfg,
		gitPath: gitPath,
		pollWg:  sync.WaitGroup{},
	}, nil
}

func (a *Agent) initRecipe(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	if _, err := os.Stat(a.cfg.Vcs.Destination); os.IsNotExist(err) {
		log.Printf("cloning recipe %s using reference %s\n", a.cfg.Vcs.URI, a.cfg.Vcs.Ref)
		cmd := exec.CommandContext(
			ctx, a.gitPath,
			"clone", a.cfg.Vcs.URI, a.cfg.Vcs.Destination,
			"--branch", a.cfg.Vcs.Ref,
		)
		if err = cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (a *Agent) pollRecipe(ctx context.Context, pollEventsChan chan<- struct{}) {
	for {
		select {
		case <-time.After(a.cfg.VcsPollDelay):
			log.Printf("updating recipe %s using reference %s\n", a.cfg.Vcs.URI, a.cfg.Vcs.Ref)
			cmd := exec.CommandContext(
				ctx, a.gitPath,
				"reset", "--hard", a.cfg.Vcs.Ref,
			)
			cmd.Dir = a.cfg.Vcs.Destination
			if err := cmd.Run(); err != nil {
				log.Printf("could not poll recipe. %v", err)
			}
			pollEventsChan <- struct{}{}
		case <-ctx.Done():
			log.Printf("stop polling properly...")
			a.pollWg.Done()
			close(pollEventsChan)
			return
		}
	}
}

func (a *Agent) executeRecipe(ctx context.Context, pollEventsChan <-chan struct{}) {
	// TODO: will be better to have the recipe passed.
	kernelModule := modules.NewKernelModule([]modules.KernelParameter{
		{Key: "fs/inotify/max_user_instances", Value: "129"},
	})
	for {
		select {
		case <-pollEventsChan:
			log.Printf("a new poll event just occured!")
			if ok := kernelModule.Dirty(); ok {
				_ = kernelModule.Do()
			}
		case <-ctx.Done():
			log.Printf("stop recipe execution properly...")
			return
		}
	}
}

func (a *Agent) Start(ctx context.Context) {
	log.Println("starting agent...")
	if err := a.initRecipe(ctx); err != nil {
		log.Fatalf("could not clone recipe. %v", err)
	}
	pollEventsChan := make(chan struct{})
	a.pollWg.Add(1)

	go a.pollRecipe(ctx, pollEventsChan)
	go a.executeRecipe(ctx, pollEventsChan)
	// Check if the configuration changed. If the configuration changed we need to apply it.
	// We need to store the state of the current infrastructure.
	// 	An idea is to run a verify() method that checks the actual status, encode and store the result.
	//  Next time a poll() occurs, if this operation returns a different value ; it means we have to perform the do() changes.
	// When we apply the configuration we need to put the poll on hold to avoid re-triggering a do() by mistake.
	//
	//
	// To conclude:
	//   Capability to store a state. First implementation should be in-memory. Maybe a file is required to not replay on startup.
	//   A modules (for e.g. KernelParameters) is equipped of a verify() and do() method.
	//   An orchestrator should call the right modules
	a.pollWg.Wait()
}