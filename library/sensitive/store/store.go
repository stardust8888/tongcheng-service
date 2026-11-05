package store

import "io"

const (
	FilterChannel = 128
)

type (
	Store interface {
		LoadDictPath(path ...string) error
		LoadDict(reader io.Reader) error
		ReadChan() <-chan string
		ReadString() []string
		GetAddChan() <-chan string
		GetDelChan() <-chan string
		AddWord(words ...string) error
		DelWord(words ...string) error
	}
)
