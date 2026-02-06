package feishu

import (
	"github.com/Duke1616/ecmdb/internal/event/domain"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/ecodeclub/ekit/slice"
)

func toCardInputFields(fields []domain.InputField) []card.InputField {
	return slice.Map(fields, func(idx int, src domain.InputField) card.InputField {
		return card.InputField{
			Name:     src.Name,
			Type:     card.FieldType(src.Type),
			Key:      src.Key,
			Required: src.Required,
			Options: slice.Map(src.Options, func(idx int, src domain.InputOption) card.InputOption {
				return card.InputOption{
					Label: src.Label,
					Value: src.Value,
				}
			}),
			Props: src.Props,
		}
	})
}

func toCardFields(fields []domain.Field) []card.Field {
	return slice.Map(fields, func(idx int, src domain.Field) card.Field {
		return card.Field{
			IsShort: src.IsShort,
			Tag:     src.Tag,
			Content: src.Content,
		}
	})
}

func toCardValues(values []domain.Value) []card.Value {
	return slice.Map(values, func(idx int, src domain.Value) card.Value {
		return card.Value{
			Key:   src.Key,
			Value: src.Value,
		}
	})
}
