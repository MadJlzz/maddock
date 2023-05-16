package modules

// Module is the contract that is implemented any service that configures the system.
type Module interface {
	Dirty() bool
	Do() error
}
