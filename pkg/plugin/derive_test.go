package plugin

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
	"github.com/samber/lo"
)

func TestRegistryWithDerive(t *testing.T) {
	def := NewRegistry("builtin.ssh", "SSH").
		Action("terminal", "SSH 终端",
			Icon("terminal"),
			Workspace("Web Shell", "host", CardFields("name", "ip")),
		).
		Setup(
			Derive[testHost]("host"),
		).
		Bind(Center[testHost]("host")).
		MustDefinition()

	if len(def.Schema.Models) != 2 {
		t.Fatalf("expected 2 derived models, got %d", len(def.Schema.Models))
	}
	if len(def.Schema.RelationTypes) != 4 {
		t.Fatalf("expected 4 basic relation types, got %d", len(def.Schema.RelationTypes))
	}

	foundHostGateway := lo.SomeBy(def.Schema.ModelRelations, func(r types.ModelRelation) bool {
		return r.SourceModelUID == "host" && r.TargetModelUID == "gateway" && r.RelationTypeUID == types.RelationTypeDefault
	})
	if !foundHostGateway {
		t.Fatalf("expected host -> gateway relation to be derived, got %#v", def.Schema.ModelRelations)
	}
}
