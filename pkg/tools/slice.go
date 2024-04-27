package tools

func ToMapBS[Ele any, Key comparable, Val any](elements []Ele, fn func(element Ele) (Key, Val)) map[Key]Val {
	resultMap := make(map[Key]Val)
	for _, element := range elements {
		k, v := fn(element)
		resultMap[k] = v
	}
	return resultMap
}

func ToMapS[Ele any, Key comparable](elements []Ele, fn func(element Ele) Key) map[Key][]Ele {
	resultMap := make(map[Key][]Ele)
	for _, element := range elements {
		key := fn(element)
		resultMap[key] = append(resultMap[key], element)
	}
	return resultMap
}
