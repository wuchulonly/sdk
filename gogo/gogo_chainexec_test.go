package gogo

import (
	"testing"

	neutrontemplates "github.com/chainreactors/neutron/templates"
	"github.com/chainreactors/sdk/pkg/types"
)

// Regression for the nil-pointer panic in engine.NeutronScan: when the SDK
// injects templates it populates pkg.TemplateMap directly (buildTemplateMap),
// bypassing pkg.LoadTemplates which is the only other place that builds
// pkg.ChainExec. buildChainExecutor must rebuild that executor so exploit
// scans don't call ChainExec.Execute on a nil receiver.
func TestBuildChainExecutorRegistersIDsAndChains(t *testing.T) {
	tmpls := []*types.Template{
		{Id: "entry", Chains: []string{"leaf"}},
		{Id: "leaf"},
		{Id: ""}, // no id: must be skipped without panicking
	}

	ce := buildChainExecutor(tmpls)
	if ce == nil {
		t.Fatal("buildChainExecutor returned nil executor")
	}
	if !ce.Has("entry") || !ce.Has("leaf") {
		t.Fatalf("registered ids = entry:%v leaf:%v, want both true", ce.Has("entry"), ce.Has("leaf"))
	}

	// "leaf" is a chain target of "entry", so only "entry" is an entrypoint.
	eps := ce.Entrypoints()
	if len(eps) != 1 || eps[0] != "entry" {
		t.Fatalf("entrypoints = %v, want [entry]", eps)
	}

	// Executing from the entrypoint must walk entry -> leaf without panic
	// (the panic being fixed is exactly this call on a nil receiver).
	var visited []string
	ce.Execute(eps, func(id string, _ map[string]interface{}) *neutrontemplates.ChainResult {
		visited = append(visited, id)
		return &neutrontemplates.ChainResult{}
	})
	if len(visited) != 2 {
		t.Fatalf("chain walk visited %v, want entry+leaf", visited)
	}
}
