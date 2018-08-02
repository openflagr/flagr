package producer

// channel based semaphore
// used to limit the number of concurrent goroutines
type semaphore chan struct{}

// acquire a lock, blocking or non-blocking
func (s semaphore) acquire() {
	s <- struct{}{}
}

// release a lock
func (s semaphore) release() {
	<-s
}

// wait block until the last goroutine release the lock
func (s semaphore) wait() {
	for i := 0; i < cap(s); i++ {
		s <- struct{}{}
	}
}
