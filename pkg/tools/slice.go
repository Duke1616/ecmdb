package tools

import "github.com/ecodeclub/ekit/slice"

func ToMap[Ele any, Key comparable](elements []Ele, fn func(element Ele) Key) map[Key]Ele {
	return ToMapV(elements, func(element Ele) (Key, Ele) {
		return fn(element), element
	})
}

func ToMapV[Ele any, Key comparable, Val any](elements []Ele, fn func(element Ele) (Key, Val)) (resultMap map[Key]Val) {
	resultMap = make(map[Key]Val, len(elements))
	for _, element := range elements {
		k, v := fn(element)
		resultMap[k] = v
	}
	return
}

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
