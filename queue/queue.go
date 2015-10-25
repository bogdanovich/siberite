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
type Item struct {
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
	}
	return q, q.open()
}

// Close leveldb database
func (q *Queue) Close() {
	if q.isOpened {
		q.db.Close()
		q.isOpened = false
	}
}

// Drop closes and deletes leveldb database
func (q *Queue) Drop() {
	q.Close()
	os.RemoveAll(q.Path())
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

// Peek returns next queue item without removing it from the queue
func (q *Queue) Peek() (*Item, error) {
	q.RLock()
	defer q.RUnlock()
	return q.GetItemByID(q.head + 1)
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

// Dequeue returns next queue item and removes it from the queue
func (q *Queue) Dequeue() (*Item, error) {
	q.Lock()
	defer q.Unlock()

	item, err := q.GetItemByID(q.head + 1)
	if err != nil {
		return item, err
	}

	err = q.db.Delete(item.Key, nil)
	if err == nil {
		q.head++
	}
	return item, err
}

// Prepend adds new item as a first element of the queue
func (q *Queue) Prepend(item *Item) error {
	q.Lock()
	defer q.Unlock()
	if q.head < 1 {
		return errors.New("queue: head can not be less then zero")
	}
	key := q.dbKey(q.head)
	err := q.db.Put(key, item.Value, nil)
	if err == nil {
		q.head--
	}
	return err
}

// GetItemByID returns an item by it's id
func (q *Queue) GetItemByID(id uint64) (*Item, error) {
	if id <= q.head || id > q.tail {
		if q.length() < 1 {
			return &Item{nil, nil}, errors.New("queue: is empty")
		}
		return &Item{nil, nil}, errors.New("queue: out of bounds")
	}

	key := q.dbKey(id)
	value, err := q.db.Get(key, nil)
	item := &Item{key, value}
	return item, err
}

// GetItemByOffset returns an item by offset from the queue head, starting from 0.
func (q *Queue) GetItemByOffset(offset uint64) (*Item, error) {
	return q.GetItemByID(q.head + 1 + offset)
}

// AddOpenTransactions increments OpenTransactions stats item
func (q *Queue) AddOpenTransactions(value int64) {
	atomic.AddInt64(&q.Stats.OpenTransactions, value)
}

// Path returns leveldb database file path
func (q *Queue) Path() string {
	return q.DataDir + "/" + q.Name
}

// GetNext returns next item value
func (q *Queue) GetNext() ([]byte, error) {
	item, err := q.Dequeue()
	return item.Value, err
}

// PutBack returns value to the queue
func (q *Queue) PutBack(value []byte) error {
	return q.Prepend(&Item{nil, value})
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

	o := opt.Options{
		BlockCacher:       opt.NoCacher,
		DisableBlockCache: true,
	}

	var err error
	q.db, err = leveldb.OpenFile(q.Path(), &o)
	if err != nil {
		return err
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
