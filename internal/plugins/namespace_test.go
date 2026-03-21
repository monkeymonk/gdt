package plugins

import "testing"

func TestResolveNamespace_ShortName_Unique(t *testing.T) {
	items := []NamespacedItem{
		{ShortName: "fps", QualifiedName: "starter:fps"},
		{ShortName: "rpg", QualifiedName: "starter:rpg"},
	}
	result, err := ResolveNamespace("fps", items)
	if err != nil {
		t.Fatal(err)
	}
	if result.QualifiedName != "starter:fps" {
		t.Errorf("got %q, want starter:fps", result.QualifiedName)
	}
}

func TestResolveNamespace_QualifiedName(t *testing.T) {
	items := []NamespacedItem{
		{ShortName: "fps", QualifiedName: "starter:fps"},
	}
	result, err := ResolveNamespace("starter:fps", items)
	if err != nil {
		t.Fatal(err)
	}
	if result.QualifiedName != "starter:fps" {
		t.Errorf("got %q, want starter:fps", result.QualifiedName)
	}
}

func TestResolveNamespace_Ambiguous(t *testing.T) {
	items := []NamespacedItem{
		{ShortName: "platformer", QualifiedName: "fps:platformer"},
		{ShortName: "platformer", QualifiedName: "rpg:platformer"},
	}
	_, err := ResolveNamespace("platformer", items)
	if err == nil {
		t.Fatal("expected ambiguity error")
	}
	ambErr, ok := err.(*AmbiguousNameError)
	if !ok {
		t.Fatalf("expected *AmbiguousNameError, got %T", err)
	}
	if len(ambErr.Candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(ambErr.Candidates))
	}
}

func TestResolveNamespace_NotFound(t *testing.T) {
	items := []NamespacedItem{
		{ShortName: "fps", QualifiedName: "starter:fps"},
	}
	_, err := ResolveNamespace("unknown", items)
	if err == nil {
		t.Fatal("expected not found error")
	}
}
