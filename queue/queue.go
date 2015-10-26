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
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Consumer represents a queue consumer
type Consumer interface {
	GetNext() ([]byte, error)
	PutBack([]byte) error
}

// Queue represents a persistent FIFO structure
// that stores the data in leveldb
type Queue struct {
	sync.RWMutex
	Name     string
	DataDir  string
	Stats    *Stats
	db       *leveldb.DB
	opts     *Options
	head     uint64
	tail     uint64
	isOpened bool
	isShared bool
}

// Options represents queue options
type Options struct {
	KeyPrefix []byte
}

// Stats contains queue level stats
type Stats struct {
	OpenTransactions int64
}

// Item represents a queue item
type item struct {
	Key   []byte
	Value []byte
}

// Open creates a queue and opens underlying leveldb database
func Open(name string, dataDir string, opts *Options) (*Queue, error) {
	q := &Queue{
		Name:     name,
		DataDir:  dataDir,
		Stats:    &Stats{0},
		db:       &leveldb.DB{},
		opts:     opts,
		head:     0,
		tail:     0,
		isOpened: false,
		isShared: false,
	}
	return q, q.open()
}

// OpenShared creates and initializes a queue from opened leveldb database
func OpenShared(name string, dataDir string,
	keyPrefix string, db *leveldb.DB) (*Queue, error) {

	q := &Queue{
		Name:     name,
		DataDir:  dataDir,
		Stats:    &Stats{0},
		db:       db,
		opts:     &Options{KeyPrefix: []byte(keyPrefix)},
		head:     0,
		tail:     0,
		isOpened: false,
		isShared: true,
	}
	return q, q.open()
}

// Close leveldb database
func (q *Queue) Close() {
	if q.isOpened && !q.isShared {
		q.db.Close()
	}
	q.isOpened = false
}

// Drop closes and deletes leveldb database
func (q *Queue) Drop() {
	q.Close()
	if !q.isShared {
		os.RemoveAll(q.Path())
	}
}

// Head returns current head offset of the queue
func (q *Queue) Head() uint64 { return q.head }

// Tail returns current tail offset of the queue
func (q *Queue) Tail() uint64 { return q.tail }

// Length returns current length of the queue
func (q *Queue) Length() uint64 {
	q.RLock()
	defer q.RUnlock()
	return q.length()
}

// Enqueue adds new value to the queue
func (q *Queue) Enqueue(value []byte) error {
	q.Lock()
	defer q.Unlock()

	key := q.dbKey(q.tail + 1)
	err := q.db.Put(key, value, nil)
	if err == nil {
		q.tail++
	}
	return err
}

// GetNext returns next value from queue
func (q *Queue) GetNext() ([]byte, error) {
	q.Lock()
	defer q.Unlock()

	item, err := q.readValueByID(q.head + 1)
	if err != nil {
		return item.Value, err
	}

	err = q.db.Delete(item.Key, nil)
	if err == nil {
		q.head++
	}
	return item.Value, err
}

// PutBack returns value to the queue
func (q *Queue) PutBack(value []byte) error {
	q.Lock()
	defer q.Unlock()
	if q.head < 1 {
		return errors.New("queue: head can not be less then zero")
	}
	key := q.dbKey(q.head)
	err := q.db.Put(key, value, nil)
	if err == nil {
		q.head--
	}
	return err
}

// Peek returns next value without removing it from the queue
func (q *Queue) Peek() ([]byte, error) {
	q.RLock()
	defer q.RUnlock()
	return q.ReadValueByID(q.head + 1)
}

// ReadValueByID returns a value by it's id
func (q *Queue) ReadValueByID(id uint64) ([]byte, error) {
	item, err := q.readValueByID(id)
	return item.Value, err
}

func (q *Queue) readValueByID(id uint64) (*item, error) {
	if id <= q.head || id > q.tail {
		if q.length() < 1 {
			return &item{nil, nil}, errors.New("queue: is empty")
		}
		return &item{nil, nil}, errors.New("queue: out of bounds")
	}

	key := q.dbKey(id)
	value, err := q.db.Get(key, nil)
	item := &item{key, value}
	return item, err
}

// ReadValueByOffset returns an item by offset from the queue head, starting from 0.
func (q *Queue) ReadValueByOffset(offset uint64) ([]byte, error) {
	return q.ReadValueByID(q.head + 1 + offset)
}

// AddOpenTransactions increments OpenTransactions stats item
func (q *Queue) AddOpenTransactions(value int64) {
	atomic.AddInt64(&q.Stats.OpenTransactions, value)
}

// Path returns leveldb database file path
func (q *Queue) Path() string {
	return q.DataDir + "/" + q.Name
}

func (q *Queue) open() error {
	q.Lock()
	defer q.Unlock()
	if regexp.MustCompile(`[^a-zA-Z0-9_]+`).MatchString(q.Name) {
		return errors.New("queue: name is not alphanumeric")
	}

	if len(q.Name) > 100 {
		return errors.New("queue: name is too long")
	}

	if !q.isShared {
		var err error
		q.db, err = leveldb.OpenFile(q.Path(), &opt.Options{})
		if err != nil {
			return err
		}
	}
	q.isOpened = true
	return q.initialize()
}

func (q *Queue) dbKey(id uint64) []byte {
	if len(q.opts.KeyPrefix) == 0 {
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, id)
		return key
	}
	key := make([]byte, len(q.opts.KeyPrefix)+8)
	copy(key[0:len(q.opts.KeyPrefix)], q.opts.KeyPrefix)
	binary.BigEndian.PutUint64(key[len(q.opts.KeyPrefix):], id)
	return key

}

func (q *Queue) dbKeyToID(key []byte) uint64 {
	return binary.BigEndian.Uint64(key[len(q.opts.KeyPrefix):])
}

func (q *Queue) length() uint64 {
	return q.tail - q.head
}

func (q *Queue) initialize() error {
	iter := q.db.NewIterator(util.BytesPrefix(q.opts.KeyPrefix), nil)
	defer iter.Release()

	if iter.First() {
		q.head = q.dbKeyToID(iter.Key()) - 1
	}

	if iter.Last() {
		q.tail = q.dbKeyToID(iter.Key())
	}

	return iter.Error()
}
