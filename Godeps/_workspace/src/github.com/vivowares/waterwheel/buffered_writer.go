package waterwheel

import (
	"bufio"
	"io"
)

type BufferedWriteCloser struct {
	size int
	cur  int
	wc   io.WriteCloser
	bw   *bufio.Writer
}

func NewBufferedWriteCloser(size int, w io.WriteCloser) *BufferedWriteCloser {
	return &BufferedWriteCloser{
		size: size,
		wc:   w,
		bw:   bufio.NewWriter(w),
	}
}

func (w *BufferedWriteCloser) Write(p []byte) (int, error) {
	if w.cur < w.size {
		w.cur += 1
		return w.bw.Write(p)
	}

	w.bw.Flush()
	w.cur = 1
	return w.bw.Write(p)
}

func (w *BufferedWriteCloser) Close() error {
	w.bw.Flush()
	return w.wc.Close()
}
