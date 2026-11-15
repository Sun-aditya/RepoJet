package facts

import (
	"testing"
)

func TestFactsNewAndAddWarning(t *testing.T) {
	f := New()

	if f.Scripts == nil {
		t.Fatal("expected Scripts map to be initialized, got nil")
	}
	if f.EnvFiles == nil {
		t.Fatal("expected EnvFiles slice to be initialized, got nil")
	}
	if f.Warnings == nil {
		t.Fatal("expected Warnings slice to be initialized, got nil")
	}

	// Empty warning should not be appended
	f.AddWarning("")
	if len(f.Warnings) != 0 {
		t.Errorf("expected 0 warnings after empty warning, got %d", len(f.Warnings))
	}

	f.AddWarning("first warning")
	f.AddWarning("second warning")

	if len(f.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(f.Warnings))
	}
	if f.Warnings[0] != "first warning" || f.Warnings[1] != "second warning" {
		t.Errorf("unexpected warnings: %v", f.Warnings)
	}
}

func TestFactStructure(t *testing.T) {
	fact := Fact{
		Value:      "Node.js",
		Evidence:   "package.json",
		Confidence: ConfidenceHigh,
	}

	if fact.Value != "Node.js" {
		t.Errorf("Value = %q, want %q", fact.Value, "Node.js")
	}
	if fact.Evidence != "package.json" {
		t.Errorf("Evidence = %q, want %q", fact.Evidence, "package.json")
	}
	if fact.Confidence != ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", fact.Confidence, ConfidenceHigh)
	}
}
