package domain

import "testing"

func TestModelRelationValidateMapping(t *testing.T) {
	tests := []struct {
		name    string
		mapping string
		wantErr bool
	}{
		{name: "empty mapping", mapping: "", wantErr: false},
		{name: "one to one", mapping: MappingOneToOne, wantErr: false},
		{name: "one to many", mapping: MappingOneToMany, wantErr: false},
		{name: "many to many", mapping: MappingManyToMany, wantErr: false},
		{name: "invalid", mapping: "many_to_one", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := ModelRelation{
				SourceModelUID:  "host",
				TargetModelUID:  "app",
				RelationTypeUID: "run",
				Mapping:         tt.mapping,
			}
			err := mr.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
