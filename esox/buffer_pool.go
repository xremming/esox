package esox

import (
	"bytes"
	"sync"
)

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

func getBytesBuffer() *bytes.Buffer {
	out := bytesBufferPool.Get().(*bytes.Buffer)
	out.Reset()
	return out
}

func putBytesBuffer(b *bytes.Buffer) {
	bytesBufferPool.Put(b)
}
