package module

import (
	"encoding/base64"
	"fmt"
	"github.com/MadJlzz/maddock/internal/state"
	"golang.org/x/exp/slices"
	"log"
)

type KernelParameter struct {
	Key   string
	Value string
}

func (kp *KernelParameter) Base64Encode() string {
	b := []byte(fmt.Sprintf("%s:%s", kp.Key, kp.Value))
	return base64.StdEncoding.EncodeToString(b)
}

type KernelModule struct {
	// StateService ==> for being able to retrieve the state when needed.
	stateService *state.StateService
	parameters   []KernelParameter
}

func NewKernelModule(parameters []KernelParameter) *KernelModule {
	// TODO: if we don't pass the state service, we could run the agent on memory automatically.
	return &KernelModule{
		stateService: state.MemStateService,
		parameters:   parameters,
	}
}

func (k *KernelModule) StateChanged() bool {
	// Load the hashes from the state.
	stateHashes := k.stateService.Get("kernel_module")
	if len(stateHashes) != len(k.parameters) {
		return true
	}
	for _, p := range k.parameters {
		if ok := slices.Contains(stateHashes, p.Base64Encode()); !ok {
			return true
		}
	}
	return false
}

func (k *KernelModule) Do() error {
	log.Println("Module state has changed, rerunning everything.")

	// Foreach KernelParameters, we apply it and insert it to the state.
	for _, p := range k.parameters {
		kp := KernelParameter{
			Key:   p.Key,
			Value: p.Value,
		}
		k.stateService.Insert("kernel_module", kp.Base64Encode())

	}
	return nil
}
