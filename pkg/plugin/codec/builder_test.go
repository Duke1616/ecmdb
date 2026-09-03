package codec

import (
	"testing"

	"github.com/Duke1616/ecmdb/pkg/plugin/types"
)

type testSubGateway struct {
	IP    string `plugin:"ip,required"`
	Level int    `plugin:"level"`
}

type testGateway struct {
	IP          string           `plugin:"ip,required"`
	Port        int              `plugin:"port,field=ssh_port"`
	SubGateways []testSubGateway `plugin:"sub_gateways,model=sub_gateway,in=run"`
}

type testHost struct {
	IP       string        `plugin:"ip,required"`
	Username string        `plugin:"username"`
	Gateways []testGateway `plugin:"gateways,model=gateway,out=default"`
}

type testInputs struct {
	Target testHost `plugin:"target"`
}

func TestBuildSpecs(t *testing.T) {
	specs, err := BuildSpecs[testInputs]()
	if err != nil {
		t.Fatalf("BuildSpecs failed: %v", err)
	}

	if len(specs) != 1 {
		t.Fatalf("expected 1 top level spec, got %d", len(specs))
	}

	target := specs[0]
	if target.Name != "target" {
		t.Errorf("expected target name 'target', got '%s'", target.Name)
	}
	if target.Cardinality != types.CardinalityOne {
		t.Errorf("expected target cardinality '%s', got '%s'", types.CardinalityOne, target.Cardinality)
	}
	if target.Fields["ip"] != "ip" || target.Fields["username"] != "username" {
		t.Errorf("unexpected target fields mapping: %v", target.Fields)
	}
	if len(target.RequiredFields) != 1 || target.RequiredFields[0] != "ip" {
		t.Errorf("unexpected required fields: %v", target.RequiredFields)
	}

	if len(target.Children) != 1 {
		t.Fatalf("expected 1 child in target, got %d", len(target.Children))
	}
	gateways := target.Children[0]
	if gateways.Name != "gateways" {
		t.Errorf("expected child name 'gateways', got '%s'", gateways.Name)
	}
	if gateways.Cardinality != types.CardinalityMany {
		t.Errorf("expected gateways cardinality '%s', got '%s'", types.CardinalityMany, gateways.Cardinality)
	}
	if gateways.ModelUID != "gateway" || gateways.RelationType != types.RelationTypeDefault {
		t.Errorf("unexpected gateway association: model=%s, relation_type=%s", gateways.ModelUID, gateways.RelationType)
	}
	if gateways.Direction != types.DirectionToTarget {
		t.Errorf("unexpected gateway direction: %s", gateways.Direction)
	}

	if len(gateways.Children) != 1 {
		t.Fatalf("expected 1 child in gateways, got %d", len(gateways.Children))
	}
	subGateways := gateways.Children[0]
	if subGateways.Name != "sub_gateways" {
		t.Errorf("expected child name 'sub_gateways', got '%s'", subGateways.Name)
	}
	if subGateways.ModelUID != "sub_gateway" || subGateways.RelationType != types.RelationTypeRun {
		t.Errorf("unexpected sub_gateway association: model=%s, relation_type=%s", subGateways.ModelUID, subGateways.RelationType)
	}
	if subGateways.Direction != types.DirectionToSource {
		t.Errorf("unexpected sub_gateway direction: %s", subGateways.Direction)
	}
}

func TestBuildCenterSpec(t *testing.T) {
	spec, err := BuildCenterSpec[testHost]("target", "host")
	if err != nil {
		t.Fatalf("BuildCenterSpec failed: %v", err)
	}
	if spec.Name != "target" || spec.ModelUID != "host" {
		t.Fatalf("unexpected center spec: %#v", spec)
	}
	if spec.Fields["ip"] != "ip" || spec.Fields["username"] != "username" {
		t.Fatalf("unexpected center fields: %v", spec.Fields)
	}
	if len(spec.Children) != 1 || spec.Children[0].Name != "gateways" {
		t.Fatalf("unexpected children: %#v", spec.Children)
	}
}
