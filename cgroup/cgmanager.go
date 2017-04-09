package cgroup

import (
	"strings"
	"sync"

	"github.com/orcaman/concurrent-map"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/bogdanovich/siberite/queue"
)

// CGManager represents multiple consumer group manager
type CGManager struct {
	cmap        cmap.ConcurrentMap
	storage     *leveldb.DB
	storagePath string
	source      *queue.Queue
	sync.Mutex
}

// NewCGManager initializes new consumer group manager
func NewCGManager(storagePath string,
	source *queue.Queue) (*CGManager, error) {

	m := &CGManager{cmap: cmap.New(), storagePath: storagePath, source: source}
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
		if cg, ok = m.get(name); !ok {
			var err error
			cg, err = NewConsumerGroup(name, m.source, m.storage)
			if err != nil {
				return nil, err
			}
			m.cmap.Set(name, cg)
		}
	}
	return cg, nil
}

// DeleteConsumerGroup deletes specified consumer group
func (m *CGManager) DeleteConsumerGroup(name string) error {
	cg, ok := m.get(name)
	if !ok {
		return nil
	}
	err := cg.Delete()
	if err != nil {
		return err
	}
	m.cmap.Remove(name)
	return nil
}

// ConsumerGroupIterator iterates through existing consumer groups
func (m *CGManager) ConsumerGroupIterator() <-chan cmap.Tuple {
	return m.cmap.IterBuffered()
}

// Close consumer group manager
func (m *CGManager) Close() {
	m.storage.Close()
	m.cmap = nil
}

func (m *CGManager) get(key string) (*ConsumerGroup, bool) {
	val, ok := m.cmap.Get(key)
	if ok {
		return val.(*ConsumerGroup), ok
	}
	return nil, ok
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
