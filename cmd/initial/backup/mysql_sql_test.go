package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGenerateInsertSQL(t *testing.T) {
	provider := &MySQLBackupProvider{}

	// 测试基本数据类型
	rowData := bson.M{
		"id":          844,
		"name":        "test",
		"status":      true,
		"score":       95.5,
		"description": nil,
		"create_time": "2025-09-26 16:13:20",
	}

	sql := provider.generateInsertSQL("test_table", rowData)

	// 验证 SQL 语句包含正确的表名
	assert.Contains(t, sql, "INSERT INTO test_table")

	// 验证包含所有字段
	assert.Contains(t, sql, "id")
	assert.Contains(t, sql, "name")
	assert.Contains(t, sql, "status")
	assert.Contains(t, sql, "score")
	assert.Contains(t, sql, "description")
	assert.Contains(t, sql, "create_time")

	// 验证值处理
	assert.Contains(t, sql, "844")                   // int
	assert.Contains(t, sql, "'test'")                // string
	assert.Contains(t, sql, "1")                     // bool true
	assert.Contains(t, sql, "95.5")                  // float
	assert.Contains(t, sql, "NULL")                  // nil
	assert.Contains(t, sql, "'2025-09-26 16:13:20'") // time string
}

func TestGenerateInsertSQLWithSpecialChars(t *testing.T) {
	provider := &MySQLBackupProvider{}

	// 测试包含特殊字符的字符串
	rowData := bson.M{
		"name": "test's value with 'quotes'",
		"desc": "normal description",
	}

	sql := provider.generateInsertSQL("test_table", rowData)

	// 验证单引号被正确转义
	assert.Contains(t, sql, "'test''s value with ''quotes'''")
	assert.Contains(t, sql, "'normal description'")
}

func TestGenerateInsertSQLWithDifferentTypes(t *testing.T) {
	provider := &MySQLBackupProvider{}

	// 测试各种数据类型
	rowData := bson.M{
		"int_val":     int32(100),
		"int64_val":   int64(200),
		"float_val":   float32(3.14),
		"float64_val": float64(2.718),
		"bool_true":   true,
		"bool_false":  false,
		"string_val":  "hello world",
		"nil_val":     nil,
	}

	sql := provider.generateInsertSQL("test_table", rowData)

	// 验证各种类型的值
	assert.Contains(t, sql, "100")
	assert.Contains(t, sql, "200")
	assert.Contains(t, sql, "3.14")
	assert.Contains(t, sql, "2.718")
	assert.Contains(t, sql, "1") // bool true
	assert.Contains(t, sql, "0") // bool false
	assert.Contains(t, sql, "'hello world'")
	assert.Contains(t, sql, "NULL")
}
