package pool

import "time"

// Option represents an option for the pool.
type Option func(*pool)

// WithResultCallback sets the result callback for the pool.
func WithResultCallback(callback func(interface{})) Option {
	return func(p *pool) {
		p.resultCallback = callback
	}
}

// WithErrorCallback sets the error callback for the pool.
func WithErrorCallback(callback func(error)) Option {
	return func(p *pool) {
		p.errorCallback = callback
	}
}

// WithRetryCount sets the retry count for the pool.
func WithRetryCount(retryCount int) Option {
	return func(p *pool) {
		p.maxRetries = retryCount
	}
}

// WithBaseDelay sets the base delay for the pool.
func WithBaseDelay(baseD time.Duration) Option {
	return func(p *pool) {
		p.baseDelay = baseD
	}
}

// WithTaskBufferSize sets the size of the Task queue for the pool.
func WithTaskBufferSize(size int) Option {
	return func(p *pool) {
		p.taskBufferSize = size
	}
}
