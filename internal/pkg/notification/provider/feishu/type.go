package feishu

import (
	"github.com/Duke1616/ecmdb/internal/pkg/notification"
	"github.com/Duke1616/enotify/notify/feishu/card"
	"github.com/ecodeclub/ekit/slice"
)

func toCardInputFields(fields []notification.InputField) []card.InputField {
	return slice.Map(fields, func(idx int, src notification.InputField) card.InputField {
		return card.InputField{
			Name:     src.Name,
			Type:     card.FieldType(src.Type),
			Key:      src.Key,
			Required: src.Required,
			Options: slice.Map(src.Options, func(idx int, src notification.InputOption) card.InputOption {
				return card.InputOption{
					Label: src.Label,
					Value: src.Value,
				}
			}),
			Props: src.Props,
		}
	})
}

func toCardFields(fields []notification.Field) []card.Field {
	return slice.Map(fields, func(idx int, src notification.Field) card.Field {
		return card.Field{
			IsShort: src.IsShort,
			Tag:     src.Tag,
			Content: src.Content,
		}
	})
}

func toCardValues(values []notification.Value) []card.Value {
	return slice.Map(values, func(idx int, src notification.Value) card.Value {
		return card.Value{
			Key:   src.Key,
			Value: src.Value,
		}
	})
}
