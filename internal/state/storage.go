package state

// This section is about backends to store the states.
// It can be done in a database ; in files but first ; we'll do that in memory.

type StateStorage interface {
	Get(moduleName string) []string
	Insert(moduleName string, hash string)
}

type InMemory struct {
	// Will probably need some mutex locks in some sort.
	data map[string][]string
}

func (im *InMemory) Get(moduleName string) []string {
	if _, exists := im.data[moduleName]; !exists {
		return []string{}
	}
	return im.data[moduleName]
}

func (im *InMemory) Insert(moduleName string, hash string) {
	if value, exists := im.data[moduleName]; exists {
		value = append(value, hash)
	} else {
		im.data[moduleName] = []string{hash}
	}
}
