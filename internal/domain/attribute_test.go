package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveDisplayFields_ExplicitDisplays(t *testing.T) {
	attrs := []Attribute{
		{ID: 1, FieldUid: "ip", FieldName: "IP地址", Display: true, Index: 2},
		{ID: 2, FieldUid: "name", FieldName: "资产名称", Display: true, Index: 1},
		{ID: 3, FieldUid: "desc", FieldName: "描述", Display: false, Index: 3},
	}

	result := ResolveDisplayFields(attrs)
	assert.Len(t, result, 2)
	assert.Equal(t, "name", result[0].FieldUid)
	assert.Equal(t, "ip", result[1].FieldUid)
}

func TestResolveDisplayFields_FallbackCandidates(t *testing.T) {
	attrs := []Attribute{
		{ID: 1, FieldUid: "file1", FieldName: "附件", FieldType: "file", SortKey: 100},
		{ID: 2, FieldUid: "builtin1", FieldName: "内置字段", Builtin: true, SortKey: 200},
		{ID: 3, FieldUid: "hostname", FieldName: "主机名", FieldType: "string", SortKey: 300},
		{ID: 4, FieldUid: "os", FieldName: "操作系统", FieldType: "string", SortKey: 400},
	}

	result := ResolveDisplayFields(attrs)
	assert.Len(t, result, 2)
	assert.Equal(t, "hostname", result[0].FieldUid)
	assert.Equal(t, "os", result[1].FieldUid)
}
