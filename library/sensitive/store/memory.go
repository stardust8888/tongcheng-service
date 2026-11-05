package store

import (
	"bufio"
	"io"
	"os"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type MemoryModel struct {
	store   cmap.ConcurrentMap[string, struct{}]
	addChan chan string
	delChan chan string
}

func NewMemoryModel() *MemoryModel {
	return &MemoryModel{
		store:   cmap.New[struct{}](),
		addChan: make(chan string, FilterChannel),
		delChan: make(chan string, FilterChannel),
	}
}

func (m *MemoryModel) LoadDictPath(paths ...string) error {
	for _, path := range paths {
		err := func(path string) error {
			f, err := os.Open(path)
			defer func(f *os.File) {
				_ = f.Close()
			}(f)
			if err != nil {
				return err
			}

			return m.LoadDict(f)
		}(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MemoryModel) LoadDict(reader io.Reader) error {
	buf := bufio.NewReader(reader)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		m.store.Set(string(line), struct{}{})
		m.addChan <- string(line)
	}

	return nil
}

func (m *MemoryModel) ReadChan() <-chan string {
	ch := make(chan string)

	go func() {
		for key := range m.store.Items() {
			ch <- key
		}
		close(ch)
	}()

	return ch
}

func (m *MemoryModel) ReadString() []string {
	res := make([]string, 0, m.store.Count())

	for key := range m.store.Items() {
		res = append(res, key)
	}

	return res
}

func (m *MemoryModel) GetAddChan() <-chan string {
	return m.addChan
}

func (m *MemoryModel) GetDelChan() <-chan string {
	return m.delChan
}

func (m *MemoryModel) AddWord(words ...string) error {
	for _, word := range words {
		m.store.Set(word, struct{}{})
		m.addChan <- word
	}

	return nil
}

func (m *MemoryModel) DelWord(words ...string) error {
	for _, word := range words {
		m.store.Remove(word)
		m.delChan <- word
	}

	return nil
}
