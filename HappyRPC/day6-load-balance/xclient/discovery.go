package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RandomRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

var _ Discovery = (*MultiServersDiscover)(nil)

type MultiServersDiscover struct {
	r       *rand.Rand
	mu      sync.Mutex
	servers []string
	index   int
}

func (m *MultiServersDiscover) Refresh() error {
	return nil
}

func (m *MultiServersDiscover) Update(servers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers = servers
	return nil
}

func (m *MultiServersDiscover) Get(mode SelectMode) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(m.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no servers available")
	}

	switch mode {
	case RandomSelect:
		return m.servers[m.r.Intn(n)], nil
	case RandomRobinSelect:
		s := m.servers[m.index%n]
		m.index = (m.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: invalid mode")
	}
}

func (m *MultiServersDiscover) GetAll() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	servers := make([]string, len(m.servers), len(m.servers))
	copy(servers, m.servers)
	return servers, nil
}

func NewMultiServersDiscover(servers []string) *MultiServersDiscover {
	m := &MultiServersDiscover{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	m.index = m.r.Intn(math.MaxInt32 - 1)
	return m
}
