package utils

import (
	"bytes"
	"sync"
)

var bytesBufferPool = sync.Pool{New: func() interface{} {
	return new(bytes.Buffer)
}}

func GetBytesBuffer() *bytes.Buffer {
	out := bytesBufferPool.Get().(*bytes.Buffer)
	out.Reset()
	return out
}

func PutBytesBuffer(b *bytes.Buffer) {
	bytesBufferPool.Put(b)
}
