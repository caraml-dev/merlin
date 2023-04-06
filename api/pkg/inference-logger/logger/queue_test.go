package logger

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSync(t *testing.T) {
	b := NewBatchQueue(3)
	if e := b.Put(1); e != nil {
		t.Error(e)
	}
	if e := b.Put(2, 3); e != nil {
		t.Error(e)
	}

	if e := b.Put(4); e != nil {
		assert.EqualError(t, e, ErrFullQueue.Error())
	}

	msgs := b.Get()
	if len(msgs) != 3 {
		t.Errorf("unexpected message count in batch: %v", len(msgs))
	}

	if e := b.Put(4); e != nil {
		t.Error(e)
	}

	msgs = b.Get()
	if len(msgs) != 1 {
		t.Errorf("unexpected message count in batch: %v", len(msgs))
	}

	if e := b.Close(); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}

	if e := b.Put(4); !errors.Is(e, ErrClosed) {
		t.Errorf("Unexpected error value: %v", e)
	}

	msgs = b.Get()
	if len(msgs) != 0 {
		t.Errorf("unexpected message count in batch: %v", len(msgs))
	}
	// close twice
	if e := b.Close(); !errors.Is(e, ErrClosed) {
		t.Errorf("Unexpected error value: %v", e)
	}
}

func TestLen(t *testing.T) {
	b := NewBatchQueue(0)
	if e := b.Put(2, 3); e != nil {
		t.Error(e)
	}

	if b.Len() != 2 {
		t.Errorf("Unexpected length: %d", b.Len())
	}
}

func TestLimits(t *testing.T) {
	b := NewBatchQueue(5)
	if e := b.Put(1, 2); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	if e := b.Put(3, 4, 5, 6, 7, 8); !errors.Is(e, ErrTooManyMessages) {
		t.Errorf("Unexpected error value: %v", e)
	}
	if e := b.Put(3, 4); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}

	for i := 1; i <= 4; i++ {
		msgs := b.GetMax(1)
		if len(msgs) != 1 {
			t.Errorf("Unexpected batch len: %v", len(msgs))
		}
		if msgs[0].(int) != i {
			t.Error("Unexpected element:", msgs[0])
		}
	}
	if e := b.Close(); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}

}

func TestGetMinMax(t *testing.T) {
	b := NewBatchQueue(0)

	var resCh = make(chan []interface{})
	var quit = make(chan bool)
	go func() {
		var result []interface{}
		for {
			if result = b.GetMinMax(2, 3); len(result) == 0 {
				quit <- true
				return
			}
			resCh <- result
		}
	}()

	if e := b.Put(1); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	if e := b.Put(2); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	result := <-resCh
	if len(result) != 2 {
		t.Errorf("Unexpected result: %v", result)
	}
	if e := b.Put(3, 4, 5, 6); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	result = <-resCh
	if len(result) != 3 {
		t.Errorf("Unexpected result: %v", result)
	}
	if e := b.Put(7); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	result = <-resCh
	if len(result) != 2 {
		t.Errorf("Unexpected result: %v", result)
	}
	_ = b.Close()
	<-quit
}

func TestGetMin(t *testing.T) {
	b := NewBatchQueue(0)

	var resCh = make(chan []interface{})
	var quit = make(chan bool)
	go func() {
		var result []interface{}
		for {
			if result = b.GetMin(2); len(result) == 0 {
				quit <- true
				return
			}
			resCh <- result
		}
	}()

	if e := b.Put(1); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	if e := b.Put(2); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	result := <-resCh
	if len(result) != 2 {
		t.Errorf("Unexpected result: %v", result)
	}
	_ = b.Close()
	<-quit
}

func TestGetAll(t *testing.T) {
	var b = NewBatchQueue(0)
	var quit = make(chan bool)
	go func() {
		for {
			if r := b.GetMinMax(3, 3); len(r) == 0 {
				quit <- true
				return
			}
		}
	}()

	if e := b.Put(2, 2); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	if res := b.GetAll(); len(res) != 2 {
		t.Errorf("Unexpected result: %v", res)
	}
	if e := b.Put(2, 2); e != nil {
		t.Errorf("Unexpected error value: %v", e)
	}
	_ = b.Close()
	if res := b.GetAll(); len(res) != 2 {
		t.Errorf("Unexpected result: %v", res)
	}
	<-quit
}

func TestAsync(t *testing.T) {
	test(t, NewBatchQueue(0), 4, 4, time.Millisecond*10)
	test(t, NewBatchQueue(10), 4, 4, time.Millisecond*10)
	test(t, NewBatchQueue(100), 1, 4, time.Millisecond*10)
	test(t, NewBatchQueue(100), 4, 1, time.Millisecond*10)
	test(t, NewBatchQueue(1000), 16, 16, time.Millisecond*30)
}

func test(t *testing.T, b *BatchQueue, sc, rc int, dur time.Duration) {
	exit := make(chan bool)
	var addCount, receiveCount int64

	// start add workers
	for i := 0; i < sc; i++ {
		go func(w int) {
			for {
				if e := b.Put(w); e != nil {
					exit <- true
					return
				}
				atomic.AddInt64(&addCount, 1)
			}
		}(i)
	}

	// start read workers
	for i := 0; i < rc; i++ {
		go func(w int) {
			for {
				msgs := b.Get()
				if len(msgs) == 0 {
					exit <- true
					return
				}
				atomic.AddInt64(&receiveCount, int64(len(msgs)))
			}
		}(i)
	}

	time.Sleep(dur)
	_ = b.Close()

	for i := 0; i < sc+rc; i++ {
		<-exit
	}

	if addCount != receiveCount {
		t.Errorf("Put and receive not equals: %v vs %v", addCount, receiveCount)
	}
	t.Logf("Added: %d", addCount)
	t.Logf("received: %d", receiveCount)
}

func BenchmarkAdd(b *testing.B) {
	mb := NewBatchQueue(0)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = mb.Put(true)
	}
}

func BenchmarkWait0(b *testing.B) {
	benchmarkWait(b, 0)
}
func BenchmarkWait1(b *testing.B) {
	benchmarkWait(b, 1)
}
func BenchmarkWait10(b *testing.B) {
	benchmarkWait(b, 10)
}
func BenchmarkWait100(b *testing.B) {
	benchmarkWait(b, 100)
}
func BenchmarkWait1000(b *testing.B) {
	benchmarkWait(b, 1000)
}
func benchmarkWait(b *testing.B, max int) {
	q := NewBatchQueue(1000)
	b.StopTimer()
	b.ReportAllocs()
	go func() {
		for {
			if e := q.Put(true); e != nil {
				return
			}
		}
	}()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.GetMax(max)
	}
	b.StopTimer()
	_ = q.Close()
}
