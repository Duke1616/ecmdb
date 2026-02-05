package tools

func UniqueBy[T any, K comparable](items []T, keyFunc func(T) K) []T {
	seen := make(map[K]struct{}, len(items))
	var result []T
	for _, item := range items {
		key := keyFunc(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
