package agent

type KernelParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ExecuteRecipeRequest struct {
	Name             string            `json:"name"`
	KernelParameters []KernelParameter `json:"kernel_parameters"`
}

type Agent struct {
}
