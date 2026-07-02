package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	pluginx "github.com/Duke1616/ecmdb/pkg/plugin"
)

func TestCompleteRelationSpec(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		spec     pluginx.ResourceSpec
		expected string
	}{
		{
			name: "out to target",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "gateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToTarget,
			},
			expected: "host_default_gateway",
		},
		{
			name: "in from source",
			base: "host",
			spec: pluginx.ResourceSpec{
				ModelUID:     "AuthGateway",
				RelationType: "default",
				Direction:    pluginx.DirectionToSource,
			},
			expected: "AuthGateway_default_host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildRelationName(tt.base, tt.spec)
			if err != nil {
				t.Fatalf("buildRelationName() error = %v", err)
			}
			if got != tt.expected {
				t.Fatalf("RelationName = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestResolveReturnsFriendlyMissingInputMessage(t *testing.T) {
	svc := &service{
		repo: &stubPluginRepo{
			plugin: domain.Plugin{
				UID:  "builtin.ssh",
				Name: "SSH",
				Actions: []domain.PluginActionSpec{
					{Action: "terminal", Name: "SSH 终端", UI: pluginx.UIBuiltinTerminal},
				},
			},
			bindingsByModelUID: map[string][]domain.PluginBinding{
				"host": {
					{
						PluginID: "builtin.ssh",
						ModelUID: "host",
						Enabled:  true,
						Graph:    mustCenterGraph(t, "target", "host", map[string]string{"ip": "ip", "username": "username"}, []string{"ip", "username"}),
					},
				},
			},
		},
		resolver: &inputResolver{
			resources: &stubResourceReader{
				findByID: map[int64]domain.Resource{
					1: {
						ID:       1,
						Name:     "host-01",
						ModelUID: "host",
						Data: map[string]any{
							"username": "root",
						},
					},
				},
			},
		},
	}

	_, err := svc.ResolveActionContext(context.Background(), pluginx.ResolveRequest{
		PluginID:   "builtin.ssh",
		Action:     "terminal",
		ResourceID: 1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "target.ip 不能为空") {
		t.Fatalf("expected friendly missing field message, got %v", err)
	}
}

func TestResolveActionContextReloadsTopLevelFields(t *testing.T) {
	reader := &stubResourceReader{
		findByID: map[int64]domain.Resource{
			1: {
				ID:       1,
				Name:     "host-01",
				ModelUID: "host",
				Data: map[string]any{
					"ip":       "10.0.0.8",
					"username": "root",
				},
			},
		},
	}
	svc := &service{
		repo: &stubPluginRepo{
			plugin: domain.Plugin{
				UID:  "builtin.ssh",
				Name: "SSH",
				Actions: []domain.PluginActionSpec{
					{Action: "terminal", Name: "SSH 终端", UI: pluginx.UIBuiltinTerminal},
				},
			},
			bindingsByModelUID: map[string][]domain.PluginBinding{
				"host": {
					{
						PluginID: "builtin.ssh",
						ModelUID: "host",
						Enabled:  true,
						Graph:    mustCenterGraph(t, "target", "host", map[string]string{"ip": "ip", "username": "username"}, []string{"ip", "username"}),
					},
				},
			},
		},
		resolver: &inputResolver{
			resources: reader,
		},
	}

	actionCtx, err := svc.ResolveActionContext(context.Background(), pluginx.ResolveRequest{
		PluginID:   "builtin.ssh",
		Action:     "terminal",
		ResourceID: 1,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	fields := actionCtx.Inputs["target"].Resources[0].Fields
	if got := fields["ip"]; got != "10.0.0.8" {
		t.Fatalf("expected ip field loaded from repository, got %v", got)
	}
	if got := fields["username"]; got != "root" {
		t.Fatalf("expected username field loaded from repository, got %v", got)
	}
	if len(reader.findByIDFields) < 2 {
		t.Fatalf("expected resource reader to be called twice, got %d", len(reader.findByIDFields))
	}
	lastFields := reader.findByIDFields[len(reader.findByIDFields)-1]
	if len(lastFields) != 2 || !containsAll(lastFields, "ip", "username") {
		t.Fatalf("expected reload with required fields, got %v", lastFields)
	}
}

func TestImportModelRelationsUpdatesChangedExistingRelation(t *testing.T) {
	modelRelations := &stubRelationModelService{
		existing: []domain.ModelRelation{
			{
				ID:              42,
				SourceModelUID:  "AuthGateway",
				TargetModelUID:  "host",
				RelationTypeUID: pluginx.RelationTypeDefault,
				RelationName:    "AuthGateway_default_host",
				Mapping:         pluginx.MappingOneToMany,
			},
		},
	}
	svc := &service{
		modelRelations: modelRelations,
	}

	err := svc.importModelRelations(context.Background(), []pluginx.ModelRelation{
		{
			SourceModelUID:  "AuthGateway",
			TargetModelUID:  "host",
			RelationTypeUID: pluginx.RelationTypeDefault,
			Mapping:         pluginx.MappingManyToMany,
		},
	})
	if err != nil {
		t.Fatalf("importModelRelations() error = %v", err)
	}

	if len(modelRelations.created) != 0 {
		t.Fatalf("expected no created relations, got %+v", modelRelations.created)
	}
	if len(modelRelations.updated) != 1 {
		t.Fatalf("expected 1 updated relation, got %+v", modelRelations.updated)
	}
	updated := modelRelations.updated[0]
	if updated.ID != 42 {
		t.Fatalf("expected existing relation ID to be preserved, got %+v", updated)
	}
	if updated.RelationName != "AuthGateway_default_host" {
		t.Fatalf("expected relation name completed, got %+v", updated)
	}
	if updated.Mapping != pluginx.MappingManyToMany {
		t.Fatalf("expected mapping updated to many_to_many, got %+v", updated)
	}
}

func TestImportModelRelationsSkipsUnchangedExistingRelation(t *testing.T) {
	modelRelations := &stubRelationModelService{
		existing: []domain.ModelRelation{
			{
				ID:              42,
				SourceModelUID:  "AuthGateway",
				TargetModelUID:  "host",
				RelationTypeUID: pluginx.RelationTypeDefault,
				RelationName:    "AuthGateway_default_host",
				Mapping:         pluginx.MappingManyToMany,
			},
		},
	}
	svc := &service{
		modelRelations: modelRelations,
	}

	err := svc.importModelRelations(context.Background(), []pluginx.ModelRelation{
		{
			SourceModelUID:  "AuthGateway",
			TargetModelUID:  "host",
			RelationTypeUID: pluginx.RelationTypeDefault,
			Mapping:         pluginx.MappingManyToMany,
		},
	})
	if err != nil {
		t.Fatalf("importModelRelations() error = %v", err)
	}
	if len(modelRelations.updated) != 0 {
		t.Fatalf("expected no updated relations, got %+v", modelRelations.updated)
	}
	if len(modelRelations.created) != 0 {
		t.Fatalf("expected no created relations, got %+v", modelRelations.created)
	}
}

type stubPluginRepo struct {
	plugin             domain.Plugin
	bindingsByModelUID map[string][]domain.PluginBinding
}

func (s *stubPluginRepo) UpsertPlugin(ctx context.Context, p domain.Plugin) error { return nil }
func (s *stubPluginRepo) UpsertBinding(ctx context.Context, b domain.PluginBinding) error {
	return nil
}
func (s *stubPluginRepo) GetBinding(ctx context.Context, uid string) (domain.PluginBinding, error) {
	return domain.PluginBinding{}, nil
}
func (s *stubPluginRepo) GetPlugin(ctx context.Context, uid string) (domain.Plugin, error) {
	return s.plugin, nil
}
func (s *stubPluginRepo) ListPlugins(ctx context.Context) ([]domain.Plugin, error) { return nil, nil }
func (s *stubPluginRepo) ListBindings(ctx context.Context) ([]domain.PluginBinding, error) {
	return nil, nil
}
func (s *stubPluginRepo) ListBindingsByPluginID(ctx context.Context, pluginID string) ([]domain.PluginBinding, error) {
	return nil, nil
}
func (s *stubPluginRepo) ListBindingsByPluginIDs(ctx context.Context, pluginIDs []string) ([]domain.PluginBinding, error) {
	return nil, nil
}
func (s *stubPluginRepo) ListEnabledBindingsByModelUID(ctx context.Context, modelUID string) ([]domain.PluginBinding, error) {
	return s.bindingsByModelUID[modelUID], nil
}
func (s *stubPluginRepo) ListEnabledBindingsByModelUIDs(ctx context.Context, modelUIDs []string) ([]domain.PluginBinding, error) {
	var res []domain.PluginBinding
	for _, modelUID := range modelUIDs {
		res = append(res, s.bindingsByModelUID[modelUID]...)
	}
	return res, nil
}
func (s *stubPluginRepo) UpdateBindingEnabled(ctx context.Context, uid string, enabled bool) error {
	return nil
}
func (s *stubPluginRepo) DeleteBinding(ctx context.Context, uid string) error { return nil }
func (s *stubPluginRepo) DeletePlugin(ctx context.Context, uid string) error  { return nil }

type stubResourceReader struct {
	findByID       map[int64]domain.Resource
	findByIDFields [][]string
}

func (s *stubResourceReader) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	copiedFields := append([]string(nil), fields...)
	s.findByIDFields = append(s.findByIDFields, copiedFields)

	resource := s.findByID[id]
	if len(fields) == 0 {
		resource.Data = map[string]any{}
		return resource, nil
	}

	filtered := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := resource.Data[field]; ok {
			filtered[field] = value
		}
	}
	resource.Data = filtered
	return resource, nil
}

func (s *stubResourceReader) ListResourceByIds(ctx context.Context, fields []string, ids []int64) ([]domain.Resource, error) {
	return nil, nil
}

func (s *stubResourceReader) ListResourcesWithFilters(
	ctx context.Context,
	fields []string,
	modelUID string,
	ids []int64,
	offset, limit int64,
	filterGroups []domain.FilterGroup,
) ([]domain.Resource, int64, error) {
	return nil, 0, nil
}

type stubRelationModelService struct {
	existing []domain.ModelRelation
	created  []domain.ModelRelation
	updated  []domain.ModelRelation
}

func (s *stubRelationModelService) CreateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	return 0, nil
}

func (s *stubRelationModelService) BatchCreate(ctx context.Context, relations []domain.ModelRelation) error {
	s.created = append(s.created, relations...)
	return nil
}

func (s *stubRelationModelService) DeleteModelRelation(ctx context.Context, id int64) (int64, error) {
	return 0, nil
}

func (s *stubRelationModelService) GetByRelationNames(ctx context.Context, names []string) ([]domain.ModelRelation, error) {
	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameSet[name] = struct{}{}
	}

	res := make([]domain.ModelRelation, 0, len(s.existing))
	for _, relation := range s.existing {
		if _, ok := nameSet[relation.RelationName]; ok {
			res = append(res, relation)
		}
	}
	return res, nil
}

func (s *stubRelationModelService) ListModelUidRelation(ctx context.Context, offset, limit int64, modelUid string) ([]domain.ModelRelation, int64, error) {
	return nil, 0, nil
}

func (s *stubRelationModelService) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	return 0, nil
}

func (s *stubRelationModelService) FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]domain.ModelDiagram, error) {
	return nil, nil
}

func (s *stubRelationModelService) CountByRelationTypeUID(ctx context.Context, uid string) (int64, error) {
	return 0, nil
}

func (s *stubRelationModelService) UpdateModelRelation(ctx context.Context, req domain.ModelRelation) (int64, error) {
	s.updated = append(s.updated, req)
	return 1, nil
}

func (s *stubRelationModelService) CheckBeforeDelete(ctx context.Context, modelUid string) error {
	return nil
}

func containsAll(fields []string, want ...string) bool {
	set := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		set[field] = struct{}{}
	}
	for _, field := range want {
		if _, ok := set[field]; !ok {
			return false
		}
	}
	return true
}

func mustCenterGraph(
	t *testing.T,
	name string,
	modelUID string,
	fields map[string]string,
	required []string,
) *pluginx.BindingGraph {
	t.Helper()

	graph, err := pluginx.GraphFromBindingSpecs(modelUID, []pluginx.ResourceSpec{
		{
			Name:           name,
			ModelUID:       modelUID,
			Cardinality:    pluginx.CardinalityOne,
			Required:       true,
			Fields:         fields,
			RequiredFields: required,
		},
	})
	if err != nil {
		t.Fatalf("GraphFromBindingSpecs() error = %v", err)
	}
	return graph
}
