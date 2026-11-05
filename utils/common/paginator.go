package common

const (
	DefaultPage     = 1
	DefaultOffset   = 0
	DefaultPageSize = 20
)

type PageReq struct {
	Page       int `form:"page" json:"page"`
	PageOffset int `form:"pageOffset" json:"pageOffset"`
	PageSize   int `form:"pageSize" json:"pageSize"`
}

func (obj *PageReq) BindDefault() {
	if obj.Page <= 0 {
		obj.Page = DefaultPage
	}
	if obj.PageOffset <= 0 {
		obj.PageOffset = DefaultOffset
	}
	if obj.PageSize <= 0 {
		obj.PageSize = DefaultPageSize
	}
}

type PageResult[T any] struct {
	Items     []T `json:"items"`
	Page      int `json:"page"`      // 当前页码
	PageSize  int `json:"pageSize"`  // 每页数量
	Total     int `json:"total"`     // 总记录数
	TotalPage int `json:"totalPage"` // 总页数
}

func PaginatorWithResult[T any](data []T, page, pageSize int) *PageResult[T] {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	total := len(data)
	totalPage := (total + pageSize - 1) / pageSize

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	var items []T
	if total > 0 {
		items = data[start:end]
	}

	return &PageResult[T]{
		Items:     items,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
	}
}

func Paginator[T any](data []T, page, pageSize int) []T {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	total := len(data)

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	var items []T
	if total > 0 {
		items = data[start:end]
	}

	return items
}
