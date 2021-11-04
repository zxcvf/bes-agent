package emitter

import "bes-agent/common/metric"

// Buffer is an object for storing metrics in a circular buffer.
type Buffer struct {
	buf chan metric.Metric
	// total dropped metrics
	drops int
	// total metrics added
	total int
}

// NewBuffer returns a Buffer
//   size is the maximum number of metrics that Buffer will cache. If Add is
//   called when the buffer is full, then the oldest metric(s) will be dropped.
func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf: make(chan metric.Metric, size),
	}
}

// IsEmpty returns true if Buffer is empty.
func (b *Buffer) IsEmpty() bool {
	return len(b.buf) == 0
}

// Len returns the current length of the buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// Drops returns the total number of dropped metrics that have occurred in this
// buffer since instantiation.
func (b *Buffer) Drops() int {
	return b.drops
}

// Total returns the total number of metrics that have been added to this buffer.
func (b *Buffer) Total() int {
	return b.total
}

// Add adds metrics to the buffer.
func (b *Buffer) Add(metrics ...metric.Metric) {
	for i := range metrics {
		b.total++
		select {
		case b.buf <- metrics[i]:
		default:
			b.drops++
			<-b.buf
			b.buf <- metrics[i]
		}
	}
}

// Batch returns a batch of metrics of size batchSize.
// the batch will be of maximum length batchSize. It can be less than batchSize,
// if the length of Buffer is less than batchSize.
func (b *Buffer) Batch(batchSize int) []metric.Metric {
	n := min(len(b.buf), batchSize)
	out := make([]metric.Metric, n)
	for i := 0; i < n; i++ {
		out[i] = <-b.buf
	}
	return out
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
