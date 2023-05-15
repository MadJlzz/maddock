package state

// This file will contains handy methods to fetch states in the other packages.

type StateService struct {
	backend StateStorage
}

func NewStateService(backend StateStorage) *StateService {
	return &StateService{backend: backend}
}

func (ss *StateService) Get(moduleName string) []string {
	return ss.backend.Get(moduleName)
}

func (ss *StateService) Insert(moduleName string, hash string) {
	ss.backend.Insert(moduleName, hash)
}

// TODO: most probably the key will be the module name and value a list of hashes.
var MemStateService = NewStateService(
	&InMemory{map[string][]string{"bob": []string{"test"}}},
)
