package module

import (
	"encoding/base64"
	"fmt"
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
	parameters []KernelParameter
}

func NewKernelModule(parameters []KernelParameter) *KernelModule {
	return &KernelModule{parameters: parameters}
}

func (k *KernelModule) StateChanged() bool {
	// Load the hashes from the state.
	var stateHashes []string
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
	return nil
}
