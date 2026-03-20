package engine

import (
	"context"
	"testing"
)

func TestRemoveNotInstalled(t *testing.T) {
	svc := testService(t)
	err := svc.Remove(context.Background(), "9.9.9")
	if err == nil {
		t.Error("expected error")
	}
	ae, ok := err.(*ActionableError)
	if !ok {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt ls" {
		t.Errorf("unexpected suggestion: %s", ae.Suggestion)
	}
}

func TestRemoveInstalled(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.3")
	err := svc.Remove(context.Background(), "4.3")
	if err != nil {
		t.Fatal(err)
	}
	if svc.IsInstalled("4.3") {
		t.Error("version should be removed")
	}
}

func TestRemoveLastVersionCleansDesktop(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.3")

	err := svc.Remove(context.Background(), "4.3")
	if err != nil {
		t.Fatal(err)
	}

	versions, _ := svc.List()
	if len(versions) != 0 {
		t.Errorf("expected 0 remaining versions, got %d", len(versions))
	}
}
