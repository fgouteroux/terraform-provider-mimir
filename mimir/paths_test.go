package mimir

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestRulesGroupPath(t *testing.T) {
	cases := []struct{ ns, name, want string }{
		{defaultNamespace, "my-group", "/config/v1/rules/default/my-group"},
		{defaultNamespace, "Harvest Rules", "/config/v1/rules/default/Harvest%20Rules"},
		{defaultNamespace, "a/b", "/config/v1/rules/default/a%2Fb"},
		{defaultNamespace, "50%", "/config/v1/rules/default/50%25"},
		{defaultNamespace, "a#b", "/config/v1/rules/default/a%23b"},
		{defaultNamespace, "a?b", "/config/v1/rules/default/a%3Fb"},
	}
	for _, c := range cases {
		if got := rulesGroupPath(c.ns, c.name); got != c.want {
			t.Errorf("rulesGroupPath(%q, %q) = %q, want %q", c.ns, c.name, got, c.want)
		}
	}
	if got := rulesNamespacePath("team a"); got != "/config/v1/rules/team%20a" {
		t.Errorf("rulesNamespacePath(space) = %q", got)
	}
	if got := rulesNamespacePath(defaultNamespace); got != "/config/v1/rules/default" {
		t.Errorf("rulesNamespacePath(identity) = %q", got)
	}
}

// TestNoRawRulerPath is a structural guard: the ruler-path literal may appear ONLY inside
// the two sanctioned helpers. It walks every string literal in the package — including
// package-level var/const initializers, closures, and ALL _test.go files (the round-3 defect
// class was a raw path in a test helper; AC4 requires product AND test coverage) — and fails
// on any occurrence outside rulesGroupPath/rulesNamespacePath. Only this file is exempt (it
// holds the marker constant and the helpers' expected-output literals).
func TestNoRawRulerPath(t *testing.T) {
	const marker = "/config" + "/v1" + "/rules" // split so this literal does not flag itself elsewhere
	allowed := map[string]bool{"rulesGroupPath": true, "rulesNamespacePath": true}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	for _, fn := range files {
		if fn == "paths_test.go" {
			continue
		}
		f, perr := parser.ParseFile(fset, fn, nil, 0)
		if perr != nil {
			t.Fatalf("parse %s: %v", fn, perr)
		}
		// [lo,hi) ranges of the sanctioned helper bodies; a marker literal is allowed only inside one.
		type span struct{ lo, hi token.Pos }
		var allow []span
		for _, d := range f.Decls {
			if fd, ok := d.(*ast.FuncDecl); ok && allowed[fd.Name.Name] && fd.Body != nil {
				allow = append(allow, span{fd.Body.Pos(), fd.Body.End()})
			}
		}
		inAllowed := func(p token.Pos) bool {
			for _, s := range allow {
				if p >= s.lo && p < s.hi {
					return true
				}
			}
			return false
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Contains(lit.Value, marker) && !inAllowed(lit.Pos()) {
				t.Errorf("%s: raw ruler path literal %s at %s is outside rulesGroupPath/rulesNamespacePath; use the helpers", fn, lit.Value, fset.Position(lit.Pos()))
			}
			return true
		})
	}
}

func TestBuildRuleGroupID(t *testing.T) {
	if got := buildRuleGroupID("tenant-1", "ns a", "grp%x"); got != "tenant-1/ns%20a/grp%25x" {
		t.Errorf("buildRuleGroupID(3-part) = %q, want %q", got, "tenant-1/ns%20a/grp%25x")
	}
	if got := buildRuleGroupID("", defaultNamespace, "my group"); got != "default/my%20group" {
		t.Errorf("buildRuleGroupID(2-part) = %q, want %q", got, "default/my%20group")
	}
}
