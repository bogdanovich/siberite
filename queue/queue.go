package queue

import (
	"encoding/binary"
	"errors"
	"os"
	"regexp"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	// ErrIsEmpty is returned when queue is empty
	ErrIsEmpty = errors.New("queue: is empty")

	// ErrIDOutOfBounds is returned when queue is is out of bounds
	ErrIDOutOfBounds = errors.New("queue: ID is out of bounds")

	// ErrInvalidName is returned when queue name is not valid
	ErrInvalidName = errors.New("queue: name is not alphanumeric")

	// ErrNameTooLong means that queue name is longer then allowed limit
	ErrNameTooLong = errors.New("queue: name is too long")

	// ErrInvalidHeadValue is returned when there is an attempt
	// to assign invalid queue head value
	ErrInvalidHeadValue = errors.New("queue: head can not be less then zero")

	// ErrSharedFlush means that there was an attempt to flush shared queue
	ErrSharedFlush = errors.New("queue: can't flush shared queue")
)

const levelDBOpenFilesCacheCapacity = 64

var validQueueNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_\-\:]+`)

// Consumer represents a queue consumer
type Consumer interface {
	GetNext() ([]byte, error)
	PutBack([]byte) error
	Peek() ([]byte, error)
	Flush() error
	Length() uint64
	IsEmpty() bool
	Stats() *Stats
}

// make sure Queue implements Consumer interface
var _ Consumer = (*Queue)(nil)

// Queue represents a persistent FIFO structure
// that stores the data in leveldb
type Queue struct {
	sync.RWMutex
	Name     string
	DataDir  string
	stats    *Stats
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

// Item represents a queue item
type Item struct {
	ID    uint64
	Key   []byte
	Value []byte
}

// Open creates a queue and opens underlying leveldb database
func Open(name string, dataDir string, opts *Options) (*Queue, error) {
	q := &Queue{
		Name:     name,
		DataDir:  dataDir,
		stats:    &Stats{0},
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
func OpenShared(name string, keyPrefix string, db *leveldb.DB) (*Queue, error) {

	q := &Queue{
		Name:     name,
		DataDir:  "",
		stats:    &Stats{0},
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
	if q.isShared {
		return
	}
	q.Close()
	os.RemoveAll(q.Path())
}

// Flush flushes all queue data
func (q *Queue) Flush() error {
	if q.isShared {
		return ErrSharedFlush
	}
	q.Lock()
	defer q.Unlock()
	q.Drop()
	return q.open()
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

// IsEmpty returns false if queue is empty
func (q *Queue) IsEmpty() bool {
	return q.Length() < 1
}

// Enqueue adds new value to the queue
func (q *Queue) Enqueue(value []byte) error {
	q.Lock()
	defer q.Unlock()

	err := q.db.Put(q.dbKey(q.tail+1), value, nil)
	if err == nil {
		q.tail++
	}
	return err
}

// GetNext returns next value from queue
func (q *Queue) GetNext() ([]byte, error) {
	q.Lock()
	defer q.Unlock()

	item, err := q.readItemByID(q.head + 1)
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
		return ErrInvalidHeadValue
	}
	err := q.db.Put(q.dbKey(q.head), value, nil)
	if err == nil {
		q.head--
	}
	return err
}

// Peek returns next value without removing it from the queue
func (q *Queue) Peek() ([]byte, error) {
	q.RLock()
	defer q.RUnlock()
	item, err := q.readItemByID(q.head + 1)
	return item.Value, err
}

// ReadItemByID returns a value by it's id
func (q *Queue) ReadItemByID(id uint64) (*Item, error) {
	q.RLock()
	defer q.RUnlock()
	return q.readItemByID(id)
}

func (q *Queue) readItemByID(id uint64) (*Item, error) {
	if id <= q.head || id > q.tail {
		if q.length() < 1 {
			return &Item{}, ErrIsEmpty
		}
		return &Item{}, ErrIDOutOfBounds
	}

	var err error
	item := &Item{ID: id, Key: q.dbKey(id)}
	item.Value, err = q.db.Get(item.Key, nil)
	return item, err
}

// ReadItemByOffset returns an item by offset from the queue head, starting from 0.
func (q *Queue) ReadItemByOffset(offset uint64) (*Item, error) {
	q.RLock()
	defer q.RUnlock()
	return q.readItemByID(q.head + 1 + offset)
}

// DeleteAll deletes all items from the queue.
// This is expensive operation. If you want to drop all elements,
// it's better to close the queue and leveldb folder
func (q *Queue) DeleteAll() error {
	q.Lock()
	defer q.Unlock()

	iter := q.db.NewIterator(util.BytesPrefix(q.opts.KeyPrefix), nil)
	defer iter.Release()
	var err error

	batch := new(leveldb.Batch)

	for iter.Next() {
		batch.Delete(iter.Key())
	}
	err = q.db.Write(batch, nil)
	if err != nil {
		return err
	}
	return q.initialize()
}

// Stats returns stats struct
func (q *Queue) Stats() *Stats {
	return q.stats
}

// Path returns leveldb database file path
func (q *Queue) Path() string {
	return q.DataDir + "/" + q.Name
}

func (q *Queue) open() error {
	if validQueueNameRegex.MatchString(q.Name) {
		return ErrInvalidName
	}

	if len(q.Name) > 100 {
		return ErrNameTooLong
	}

	if !q.isShared {
		var err error
		q.db, err = leveldb.OpenFile(
			q.Path(),
			&opt.Options{OpenFilesCacheCapacity: levelDBOpenFilesCacheCapacity},
		)
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
	} else {
		q.head = 0
	}

	if iter.Last() {
		q.tail = q.dbKeyToID(iter.Key())
	} else {
		q.tail = 0
	}

	return iter.Error()
}
