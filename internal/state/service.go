package state

type Service struct {
	backend Storage
}

func NewStateService(storageType StorageType) *Service {
	var backend Storage
	switch storageType {
	default:
		backend = newInMemory()
	}
	return &Service{backend: backend}
}

func (ss *Service) Get(moduleName string) []string {
	return ss.backend.Get(moduleName)
}

func (ss *Service) Insert(moduleName string, hash string) {
	ss.backend.Insert(moduleName, hash)
}

// TODO: most probably the key will be the modules name and value a list of hashes.
var MemStateService = NewStateService(InMemory)
