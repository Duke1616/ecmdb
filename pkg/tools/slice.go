package tools

import "github.com/mitchellh/mapstructure"

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

func ConvertSlice[S any, D any](source []S) ([]D, error) {
	var result []D

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &result,
		TagName:          "json",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(source); err != nil {
		return nil, err
	}

	return result, nil
}
