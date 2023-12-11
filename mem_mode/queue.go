package memmode

type queue chan *Transaction

type queuePool struct {
	pool    []queue
	current int
}

func newQueuePool(size int, sizeLimit int) *queuePool {
	pool := make([]queue, size)
	for i := 0; i < size; i++ {
		pool[i] = make(queue, sizeLimit*10)
	}
	return &queuePool{
		pool: pool,
	}
}

func (q *queuePool) getQueueRoundRobin() queue {
	q.current = (q.current + 1) % len(q.pool)
	return q.pool[q.current]
}
