package eventstore

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"sync"
)

// We use a pool of gzip.Writers to not stress the GC
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		buff := bytes.Buffer{}
		w, err := gzip.NewWriterLevel(&buff, gzip.BestSpeed)
		if err != nil {
			panic("Could not allocate writer in gzipWriterPool: " + err.Error())
		}
		return w
	},
}

// We use a pool of gzip.Readers to not stress the GC
var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		r := new(gzip.Reader)
		return r
	},
}

// MustCompress the given buffer with gzip and returns the result
func MustCompress(b []byte) []byte {
	buff := bytes.Buffer{}
	w := gzipWriterPool.Get().(*gzip.Writer)
	defer gzipWriterPool.Put(w)
	w.Reset(&buff)
	_, err := w.Write(b)
	if err != nil {
		panic("Could not compress buffer: " + err.Error())
	}
	err = w.Close()
	if err != nil {
		panic("Could not close stream for compression: " + err.Error())
	}
	return buff.Bytes()
}

// MustDecompress the given buffer with gzip and returns the result
func MustDecompress(b []byte) []byte {
	buff := bytes.NewReader(b)
	r := gzipReaderPool.Get().(*gzip.Reader)
	defer gzipReaderPool.Put(r)
	r.Reset(buff)
	result, err := ioutil.ReadAll(r)
	if err != nil {
		panic("Could not decompress buffer: " + err.Error())
	}
	return result
}
