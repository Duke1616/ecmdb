package dao

import (
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildExcludeAndFilterBsonIgnoresEmptyFilterName(t *testing.T) {
	dao := &resourceDAO{}

	filter := dao.buildExcludeAndFilterBson("host", []int64{1, 2}, domain.Condition{
		Name:      " ",
		Condition: "equal",
		Input:     "ignored",
	})

	assert.Equal(t, "host", filter["model_uid"])
	assert.Equal(t, bson.M{"$nin": []int64{1, 2}}, filter["id"])
	assert.NotContains(t, filter, "")
}

func TestBuildBsonConditionIgnoresEmptyFieldUID(t *testing.T) {
	condition := buildBsonCondition(domain.FilterCondition{
		FieldUID: " ",
		Operator: domain.OperatorEq,
		Value:    "ignored",
	})

	assert.Nil(t, condition)
}

func TestBuildProjectionIgnoresEmptyFields(t *testing.T) {
	projection := buildProjection([]string{"name", " ", "", "ip"})

	assert.Equal(t, 1, projection["name"])
	assert.Equal(t, 1, projection["ip"])
	assert.NotContains(t, projection, "")
}
