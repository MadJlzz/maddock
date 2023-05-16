package state

type StorageType int

const (
	InMemory StorageType = iota
)

type Storage interface {
	Get(moduleName string) []string
	Insert(moduleName string, hash string)
}

type inMemory struct {
	// Will probably need some mutex locks in some sort.
	data map[string][]string
}

func newInMemory() *inMemory {
	return &inMemory{
		data: map[string][]string{
			"kernel_module": {},
		},
	}
}

func (im *inMemory) Get(moduleName string) []string {
	if _, exists := im.data[moduleName]; !exists {
		return []string{}
	}
	return im.data[moduleName]
}

func (im *inMemory) Insert(moduleName string, hash string) {
	im.data[moduleName] = append(im.data[moduleName], hash)
}
