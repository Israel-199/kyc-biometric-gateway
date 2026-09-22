package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        State
	failureCount int
	maxFailures  int
	resetTimeout time.Duration
	lastStateChange time.Time
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     maxFailures,
		resetTimeout:    resetTimeout,
		lastStateChange: time.Now(),
	}
}

func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()

	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}

	cb.mu.Unlock()

	result, err := req()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
		return nil, err
	}

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failureCount = 0
		cb.lastStateChange = time.Now()
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}

	return result, nil
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
