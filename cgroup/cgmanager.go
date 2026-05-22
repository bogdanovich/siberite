package cgroup

import (
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/bogdanovich/siberite/queue"
)

// CGManager represents multiple consumer group manager
type CGManager struct {
	sync.RWMutex
	groups      map[string]*ConsumerGroup
	storage     *leveldb.DB
	storagePath string
	source      *queue.Queue
}

// ConsumerGroupItem is a snapshot item for an existing consumer group.
type ConsumerGroupItem struct {
	Key string
	Val *ConsumerGroup
}

// NewCGManager initializes new consumer group manager
func NewCGManager(storagePath string,
	source *queue.Queue) (*CGManager, error) {

	m := &CGManager{groups: make(map[string]*ConsumerGroup), storagePath: storagePath, source: source}
	var err error
	m.storage, err = leveldb.OpenFile(storagePath, &opt.Options{})
	if err != nil {
		return m, err
	}
	return m, m.initialize()
}

// ConsumerGroup returns queue interface for provided consumer group name
func (m *CGManager) ConsumerGroup(name string) (*ConsumerGroup, error) {
	cg, ok := m.get(name)
	if !ok {
		m.Lock()
		defer m.Unlock()
		if cg, ok = m.groups[name]; !ok {
			var err error
			cg, err = NewConsumerGroup(name, m.source, m.storage)
			if err != nil {
				return nil, err
			}
			m.groups[name] = cg
		}
	}
	return cg, nil
}

// DeleteConsumerGroup deletes specified consumer group
func (m *CGManager) DeleteConsumerGroup(name string) error {
	m.Lock()
	defer m.Unlock()
	cg, ok := m.groups[name]
	if !ok {
		return nil
	}
	err := cg.Delete()
	if err != nil {
		return err
	}
	delete(m.groups, name)
	return nil
}

// ConsumerGroupIterator iterates through existing consumer groups
func (m *CGManager) ConsumerGroupIterator() <-chan ConsumerGroupItem {
	items := m.consumerGroupItems()
	ch := make(chan ConsumerGroupItem, len(items))
	for _, item := range items {
		ch <- item
	}
	close(ch)
	return ch
}

// ConsumerGroups returns a snapshot of existing consumer groups.
func (m *CGManager) ConsumerGroups() []*ConsumerGroup {
	items := m.consumerGroupItems()
	groups := make([]*ConsumerGroup, 0, len(items))
	for _, item := range items {
		groups = append(groups, item.Val)
	}
	return groups
}

// Close consumer group manager
func (m *CGManager) Close() {
	m.storage.Close()
	m.Lock()
	defer m.Unlock()
	m.groups = nil
}

func (m *CGManager) get(key string) (*ConsumerGroup, bool) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.groups[key]
	return val, ok
}

func (m *CGManager) consumerGroupItems() []ConsumerGroupItem {
	m.RLock()
	defer m.RUnlock()
	items := make([]ConsumerGroupItem, 0, len(m.groups))
	for key, cg := range m.groups {
		items = append(items, ConsumerGroupItem{Key: key, Val: cg})
	}
	return items
}

func (m *CGManager) initialize() error {
	var (
		err    error
		cgName string
	)

	iter := m.storage.NewIterator(util.BytesPrefix([]byte(cgCursorPrefix)), nil)
	defer iter.Release()

	for iter.Next() {
		cgName = strings.TrimPrefix(string(iter.Key()), cgCursorPrefix)
		_, err = m.ConsumerGroup(cgName)
		if err != nil {
			return err
		}
	}
	return iter.Error()
}
