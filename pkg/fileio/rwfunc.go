package fileio

import "io"

func ReadCloser(read bytesFunc) io.ReadCloser    { return rwFunc(read) }
func Reader(read bytesFunc) io.Reader            { return rwFunc(read) }
func WriteCloser(write bytesFunc) io.WriteCloser { return rwFunc(write) }
func Writer(write bytesFunc) io.Writer           { return rwFunc(write) }

type bytesFunc = func(b []byte) (n int, err error)

type rwFunc bytesFunc

func (f rwFunc) Read(b []byte) (n int, err error)  { return f(b) }
func (f rwFunc) Write(b []byte) (n int, err error) { return f(b) }
func (f rwFunc) Close() (err error)                { return }
