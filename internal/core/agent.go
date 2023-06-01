package core

import (
	"context"
	"fmt"
	"github.com/MadJlzz/maddock/internal/recipe"
	"gopkg.in/yaml.v3"
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
			cmd := exec.CommandContext(ctx, a.gitPath, "fetch", "origin")
			cmd.Dir = a.cfg.Vcs.Destination
			if err := cmd.Run(); err != nil {
				log.Printf("could not fetch recipes. %v", err)
			}
			log.Printf("reseting recipes %s using reference %s\n", a.cfg.Vcs.URI, a.cfg.Vcs.Ref)
			cmd = exec.CommandContext(
				ctx, a.gitPath,
				"reset", "--hard", fmt.Sprintf("origin/%s", a.cfg.Vcs.Ref),
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
	for {
		select {
		case <-pollEventsChan:
			log.Printf("a new poll event just occured!")

			recipesFilepath := recipe.DiscoverRecipes(a.cfg.Vcs.Destination)

			// TODO: better to merge all recipes found into one big recipe.
			// It might be something like a merge of multiple YAML files.
			// Then there is the unmarshal part. I guess this will me moved in the parser section.
			var r recipe.Recipe
			fd, _ := os.ReadFile(recipesFilepath[0])
			if err := yaml.Unmarshal(fd, &r); err != nil {
				panic(err)
			}

			// TODO: same here; we need to execute each modules.
			// Even here, some async execution would be beneficial.
			// Have to think about modules that could have concurrent race conditions.
			if ok := r.Modules[0].Dirty(); ok {
				_ = r.Modules[0].Do()
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
	go a.pollRecipe(ctx, pollEventsChan)
	go a.executeRecipe(ctx, pollEventsChan)

	a.pollWg.Wait()
}
