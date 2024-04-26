package tools

import "github.com/ecodeclub/ekit/slice"

func ToMapS[Ele any, Key comparable](elements []Ele, fn func(element Ele) Key) map[Key][]Ele {
	resultMap := make(map[Key][]Ele)
	for _, element := range elements {
		key := fn(element)
		resultMap[key] = append(resultMap[key], element)
	}
	return resultMap
}

func ToMapSlice[Ele any, Key comparable](elements []Ele, fn func(element Ele) Key) map[Key][]Ele {
	return slice.ToMapV(elements, func(element Ele) (Key, []Ele) {
		var result []Ele
		for _, ele := range elements {
			// TODO key 过滤
			result = append(result, ele)
		}

		return fn(element), result
	})
}
