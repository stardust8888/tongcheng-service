package treex

type Node[T any, K comparable] struct {
	ID       K             `json:"id"`
	Data     T             `json:"data"`
	Children []*Node[T, K] `json:"children,omitempty"`
}

// BuildTree 平级切片 -> 嵌套树
func BuildTree[T any, K comparable](flat []T, getID func(T) K, getPID func(T) *K) []*Node[T, K] {
	nodeMap := make(map[K]*Node[T, K], len(flat))
	var roots []*Node[T, K]

	// 第一轮：造节点
	for _, v := range flat {
		nodeMap[getID(v)] = &Node[T, K]{
			Data:     v,
			ID:       getID(v),
			Children: nil,
		}
	}

	// 第二轮：挂父子
	for _, v := range flat {
		n := nodeMap[getID(v)]
		if p := getPID(v); p != nil {
			if parent, ok := nodeMap[*p]; ok {
				parent.Children = append(parent.Children, n)
				continue
			}
			roots = append(roots, n)
			continue
		}

		roots = append(roots, n)
	}

	return roots
}
