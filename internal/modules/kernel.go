package modules

import (
	"encoding/base64"
	"fmt"
	"github.com/MadJlzz/maddock/internal/state"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
)

type KernelParameter struct {
	Key   string
	Value string
}

func (kp *KernelParameter) String() string {
	return fmt.Sprintf("%s:%s", kp.Key, kp.Value)
}

func (kp *KernelParameter) UnmarshalYAML(value *yaml.Node) error {
	kp.Key = value.Content[0].Value
	kp.Value = value.Content[1].Value
	return nil
}

type KernelParameterModule struct {
	stateService *state.Service
	parameters   []KernelParameter
}

func NewKernelModule(parameters []KernelParameter) *KernelParameterModule {
	// TODO: if we don't pass the state core, we could run the agent on memory automatically.
	return &KernelParameterModule{
		stateService: state.MemStateService,
		parameters:   parameters,
	}
}

func (k *KernelParameterModule) Base64Encode() string {
	stringifyParameters := make([]string, len(k.parameters))
	for _, v := range k.parameters {
		stringifyParameters = append(stringifyParameters, v.String())
	}
	b := []byte(strings.Join(stringifyParameters, "."))
	return base64.StdEncoding.EncodeToString(b)
}

func (k *KernelParameterModule) Dirty() bool {
	stateHash := k.stateService.Get("kernel_module")
	if stateHash != k.Base64Encode() {
		return true
	}
	return false
}

func (k *KernelParameterModule) Do() error {
	log.Println("Module state has changed, rerunning everything.")
	// Foreach KernelParameters, we apply it and insert it to the state.
	for _, p := range k.parameters {
		// TODO: compute the whole string to encode.
		kp := KernelParameter{
			Key:   p.Key,
			Value: p.Value,
		}
		fmt.Println(kp)
	}
	k.stateService.Insert("kernel_module", k.Base64Encode())
	return nil
}

func (k *KernelParameterModule) UnmarshalYAML(value *yaml.Node) error {
	var kp []KernelParameter
	if err := value.Decode(&kp); err != nil {
		return err
	}

	tmpKernelModule := NewKernelModule(kp)
	k.parameters = tmpKernelModule.parameters
	// TODO: This value should be configured from the actual agent configuration.
	k.stateService = tmpKernelModule.stateService

	return nil
}
