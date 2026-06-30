package plugin

import "testing"

type testEndpoint struct {
	Host     string `plugin:"host,required"`
	Port     int    `plugin:"port,default=22"`
	Username string `plugin:"username,required"`
}

type testTarget struct {
	testEndpoint
	Gateways []testEndpoint `plugin:"gateways"`
}

func TestInputOne(t *testing.T) {
	actionCtx := ActionContext{
		Inputs: map[string]ResolvedInput{
			"target": {
				Name:        "target",
				Cardinality: CardinalityOne,
				Resources: []ResolvedResource{
					{
						Fields: map[string]any{
							"host":     "10.0.0.1",
							"username": "root",
						},
						Children: map[string]ResolvedInput{
							"gateways": {
								Name:        "gateways",
								Cardinality: CardinalityMany,
								Resources: []ResolvedResource{
									{
										Fields: map[string]any{
											"host":     "10.0.0.10",
											"port":     2222,
											"username": "jump",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	target, err := InputOne[testTarget](actionCtx, "target")
	if err != nil {
		t.Fatalf("InputOne() error = %v", err)
	}
	if target.Host != "10.0.0.1" || target.Port != 22 || target.Username != "root" {
		t.Fatalf("target = %#v", target)
	}
	if len(target.Gateways) != 1 {
		t.Fatalf("len(gateways) = %d", len(target.Gateways))
	}
	if target.Gateways[0].Host != "10.0.0.10" || target.Gateways[0].Port != 2222 {
		t.Fatalf("gateway = %#v", target.Gateways[0])
	}
}
