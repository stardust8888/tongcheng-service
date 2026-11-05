package sensitive

import (
	"tongcheng/library/sensitive/filter"
	"tongcheng/library/sensitive/store"
)

const (
	StoreMemory = iota
)

const (
	FilterDfa = iota
)

type StoreOption struct {
	Type uint32
}

type FilterOption struct {
	Type uint32
}

type Manager struct {
	store.Store
	filter.Filter
}

func NewFilter(storeOption StoreOption, filterOption FilterOption) *Manager {
	var filterStore store.Store
	var myFilter filter.Filter

	switch storeOption.Type {
	case StoreMemory:
		filterStore = store.NewMemoryModel()
	default:
		panic("invalid store type")
	}

	switch filterOption.Type {
	case FilterDfa:
		dfaModel := filter.NewDfaModel()

		go dfaModel.Listen(filterStore.GetAddChan(), filterStore.GetDelChan())

		myFilter = dfaModel
	default:
		panic("invalid filter type")
	}

	return &Manager{
		Store:  filterStore,
		Filter: myFilter,
	}
}

var (
	manager  *Manager
	dictPath = []string{"./dict/default.txt"}
)

func Init() error {
	manager = NewFilter(StoreOption{
		Type: StoreMemory,
	}, FilterOption{
		Type: FilterDfa,
	})
	err := manager.LoadDictPath(dictPath...)
	if err != nil {
		return err
	}

	return nil
}

func GetSenFilter() *Manager {
	return manager
}
