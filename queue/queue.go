package queue

import (
	"encoding/binary"
	"errors"
	"os"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Queue struct {
	sync.RWMutex
	Name     string
	DataDir  string
	Stats    *Stats
	head     uint64
	tail     uint64
	db       *leveldb.DB
	isOpened bool
}

type Stats struct {
	OpenTransactions int64
}

// Queue Item
type Item struct {
	Key   []byte
	Value []byte
	Size  int32
}

// Create Queue and open leveldb database
func Open(name string, dataDir string) (*Queue, error) {
	q := &Queue{
		Name:     name,
		DataDir:  dataDir,
		Stats:    &Stats{0},
		db:       &leveldb.DB{},
		head:     0,
		tail:     0,
		isOpened: false,
	}
	return q, q.open()
}

// Close leveldb database
func (self *Queue) Close() {
	if self.isOpened {
		self.db.Close()
	}
	self.isOpened = false
}

// Close and delete leveldb database
func (self *Queue) Drop() {
	self.Close()
	os.RemoveAll(self.Path())
}

// Current head of the queue
func (self *Queue) Head() uint64 { return self.head }

// Current tail of the queue
func (self *Queue) Tail() uint64 { return self.tail }

// Current length of the queue
func (self *Queue) Length() uint64 {
	self.RLock()
	defer self.RUnlock()
	return self.length()
}

// Get next item without removing it from the queue
func (self *Queue) Peek() (*Item, error) {
	self.RLock()
	defer self.RUnlock()

	return self.peek()
}

// Add new value to the queue
func (self *Queue) Enqueue(value []byte) error {
	self.Lock()
	defer self.Unlock()

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, self.tail+1)
	err := self.db.Put(key, value, nil)
	if err == nil {
		self.tail += 1
	}
	return err
}

// Get next item and remove it from queue
func (self *Queue) Dequeue() (*Item, error) {
	self.Lock()
	defer self.Unlock()

	item, err := self.peek()
	if err != nil {
		return item, err
	}

	err = self.db.Delete(item.Key, nil)
	if err == nil {
		self.head += 1
	}
	return item, err
}

// Put item as a new head item
func (self *Queue) Prepend(item *Item) error {
	self.Lock()
	defer self.Unlock()
	if self.head < 1 {
		return errors.New("Queue head can not be less then zero")
	}
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, self.head)
	err := self.db.Put(key, item.Value, nil)
	if err == nil {
		self.head -= 1
	}
	return err
}

func (self *Queue) AddOpenTransactions(value int64) {
	atomic.AddInt64(&self.Stats.OpenTransactions, value)
}

func (self *Queue) open() error {
	self.Lock()
	defer self.Unlock()
	if regexp.MustCompile(`[^a-zA-Z0-9_]+`).MatchString(self.Name) {
		return errors.New("Queue name is not alphanumeric")
	}

	if len(self.Name) > 100 {
		return errors.New("Queue name is too long")
	}

	var options opt.Options
	options.BlockCacher = opt.NoCacher

	var err error
	self.db, err = leveldb.OpenFile(self.Path(), &options)
	if err != nil {
		return err
	}
	self.isOpened = true
	return self.initialize()
}

func (self *Queue) length() uint64 {
	return self.tail - self.head
}

func (self *Queue) peek() (*Item, error) {
	if self.length() < 1 {
		return &Item{nil, nil, 0}, errors.New("Queue is empty")
	}

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, self.head+1)
	value, err := self.db.Get(key, nil)
	item := &Item{key, value, int32(len(value))}
	return item, err
}

func (self *Queue) initialize() error {
	iter := self.db.NewIterator(nil, nil)
	defer iter.Release()

	if iter.First() {
		self.head = binary.BigEndian.Uint64(iter.Key()) - 1
	}

	if iter.Last() {
		self.tail = binary.BigEndian.Uint64(iter.Key())
	}

	return iter.Error()
}

func (self *Queue) Path() string {
	return self.DataDir + "/" + self.Name
}
