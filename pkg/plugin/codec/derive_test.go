package codec

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

func TestDeriveSchema(t *testing.T) {
	schema, err := DeriveSchema[testHost]("host")
	if err != nil {
		t.Fatalf("DeriveSchema() error = %v", err)
	}

	if len(schema.Models) < 2 {
		t.Fatalf("expected at least 2 models (host and gateway), got %d", len(schema.Models))
	}

	var hostModel, gatewayModel *types.ModelSpec
	for i := range schema.Models {
		if schema.Models[i].UID == "host" {
			hostModel = &schema.Models[i]
		} else if schema.Models[i].UID == "gateway" {
			gatewayModel = &schema.Models[i]
		}
	}

	if hostModel == nil {
		t.Fatal("expected host model to be derived")
	}
	if gatewayModel == nil {
		t.Fatal("expected gateway model to be derived")
	}

	hostFields := hostModel.AttributeGroups[0].Fields
	if len(hostFields) != 2 {
		t.Fatalf("expected 2 fields in host, got %d", len(hostFields))
	}

	if len(schema.Models) != 3 {
		t.Fatalf("expected 3 models (host, gateway, sub_gateway), got %d", len(schema.Models))
	}

	if len(schema.ModelRelations) != 2 {
		t.Fatalf("expected 2 model relations, got %d", len(schema.ModelRelations))
	}
	foundHostGateway := lo.SomeBy(schema.ModelRelations, func(r types.ModelRelation) bool {
		return r.SourceModelUID == "host" && r.TargetModelUID == "gateway" && r.RelationTypeUID == types.RelationTypeDefault
	})
	if !foundHostGateway {
		t.Fatalf("expected host -> gateway relation, got %#v", schema.ModelRelations)
	}
}
