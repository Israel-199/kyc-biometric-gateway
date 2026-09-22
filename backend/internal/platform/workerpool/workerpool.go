package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

type Task func(ctx context.Context) error

type WorkerPool struct {
	maxWorkers   int
	taskQueue    chan Task
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	running      int32
	activeWorker int32
}

func NewWorkerPool(maxWorkers int, queueCapacity int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		maxWorkers: maxWorkers,
		taskQueue:  make(chan Task, queueCapacity),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (wp *WorkerPool) Start() {
	if !atomic.CompareAndSwapInt32(&wp.running, 0, 1) {
		return
	}

	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}
			atomic.AddInt32(&wp.activeWorker, 1)
			_ = task(wp.ctx)
			atomic.AddInt32(&wp.activeWorker, -1)
		}
	}
}

func (wp *WorkerPool) Submit(task Task) error {
	if atomic.LoadInt32(&wp.running) == 0 {
		return errors.New("worker pool is stopped")
	}

	select {
	case wp.taskQueue <- task:
		return nil
	default:
		return errors.New("worker pool queue full")
	}
}

func (wp *WorkerPool) Stop() {
	if !atomic.CompareAndSwapInt32(&wp.running, 1, 0) {
		return
	}

	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
}

func (wp *WorkerPool) ActiveWorkers() int32 {
	return atomic.LoadInt32(&wp.activeWorker)
}
