package logger

import (
	"errors"
	"sync"
)

// ErrClosed is returned when you add message to closed queue
var ErrClosed = errors.New("BatchQueue closed")

// ErrTooManyMessages means that adding more messages (at one call) than the limit
var ErrTooManyMessages = errors.New("Too many messages")

// ErrFullQueue means that queue is full and incoming request log will be dropped
var ErrFullQueue = errors.New("Queue is full, request log is dropped")

// NewBatchQueue returns a new BatchQueue with given queue size.
// size <= 0 means unlimited
func NewBatchQueue(size int) *BatchQueue {
	return &BatchQueue{
		cond: sync.NewCond(&sync.Mutex{}),
		size: size,
		read: make(chan struct{}),
	}
}

// BatchQueue - message batching object
// Implements queue.
// Based on condition variables
type BatchQueue struct {
	messages []interface{}
	cond     *sync.Cond
	size     int
	wait     int
	read     chan struct{}

	closed bool
}

// Get until anybody add message
// Returning array of accumulated messages
// When queue will be closed length of array will be 0
func (q *BatchQueue) Get() (msgs []interface{}) {
	return q.GetMinMax(0, 0)
}

// GetMax it's Get with limit of maximum returning array size
func (q *BatchQueue) GetMax(max int) (msgs []interface{}) {
	return q.GetMinMax(0, max)
}

// GetMin it's Get with limit of minimum returning array size
func (q *BatchQueue) GetMin(min int) (msgs []interface{}) {
	return q.GetMinMax(min, 0)
}

// GetMinMax it's Get with limit of minimum and maximum returning array size
// value < 0 means no limit
func (q *BatchQueue) GetMinMax(min, max int) (msgs []interface{}) {
	if min <= 0 {
		min = 1
	}
	q.cond.L.Lock()

	// wait till number of queued message reach min threshold
	for len(q.messages) < min {
		if q.closed {
			q.cond.L.Unlock()
			return
		}
		q.cond.Wait()
	}

	if max > 0 && len(q.messages) > max {
		msgs = q.messages[:max]
		q.messages = q.messages[max:]
	} else {
		msgs = q.messages
		q.messages = make([]interface{}, 0)
	}
	q.unlockAdd()
	q.cond.L.Unlock()
	return
}

// GetAll return all messages and flush queue
// Works on closed queue
func (q *BatchQueue) GetAll() (msgs []interface{}) {
	q.cond.L.Lock()
	msgs = q.messages
	q.messages = make([]interface{}, 0)
	q.unlockAdd()
	q.cond.L.Unlock()
	return
}

// Put - adds new messages to queue.
// When queue is closed - returning ErrClosed
// When count messages bigger then queue size - returning ErrTooManyMessages
// When the queue is full - wait until will free place
func (q *BatchQueue) Put(msgs ...interface{}) (err error) {
	q.cond.L.Lock()
	// check for close
	if q.closed {
		q.cond.L.Unlock()
		return ErrClosed
	}

	for q.size > 0 && len(q.messages)+len(msgs) > q.size {
		if len(msgs) > q.size {
			q.cond.L.Unlock()
			return ErrTooManyMessages
		}

		if q.closed {
			q.cond.L.Unlock()
			return ErrClosed
		}

		q.cond.L.Unlock()
		// nolint:staticcheck
		return ErrFullQueue
	}

	q.messages = append(q.messages, msgs...)
	q.cond.L.Unlock()
	q.cond.Signal()
	return
}

func (q *BatchQueue) unlockAdd() {
	if q.wait > 0 {
		for i := 0; i < q.wait; i++ {
			q.read <- struct{}{}
		}
		q.wait = 0
	}
}

// Len returning current size of queue
func (q *BatchQueue) Len() (l int) {
	q.cond.L.Lock()
	l = len(q.messages)
	q.cond.L.Unlock()
	return
}

// Close closes the queue
// All added messages will be available for Get
// When queue paused messages do not be released for Get (use GetAll for fetching them)
func (q *BatchQueue) Close() (err error) {
	q.cond.L.Lock()
	if q.closed {
		q.cond.L.Unlock()
		return ErrClosed
	}
	q.closed = true
	q.unlockAdd()
	q.cond.L.Unlock()
	q.cond.Broadcast()
	return
}
