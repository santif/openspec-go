package commandgen

import "testing"

// mockAdapter implements ToolCommandAdapter for testing.
type mockAdapter struct {
	toolID    string
	skillsDir string
}

func (m *mockAdapter) GetToolID() string                                         { return m.toolID }
func (m *mockAdapter) GetSkillsDir() string                                      { return m.skillsDir }
func (m *mockAdapter) GenerateSkills(workflows []string, version string) []CommandContent { return nil }
func (m *mockAdapter) GenerateCommands(workflows []string) []CommandContent               { return nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	// Save and restore original registry
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	adapter := &mockAdapter{toolID: "test-tool", skillsDir: ".test"}
	Register(adapter)

	got := Get("test-tool")
	if got == nil {
		t.Fatal("expected adapter, got nil")
	}
	if got.GetToolID() != "test-tool" {
		t.Errorf("GetToolID() = %q, want %q", got.GetToolID(), "test-tool")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	got := Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for nonexistent tool, got %v", got)
	}
}

func TestRegistry_AllAdapters(t *testing.T) {
	orig := registry
	registry = map[string]ToolCommandAdapter{}
	defer func() { registry = orig }()

	Register(&mockAdapter{toolID: "tool-a"})
	Register(&mockAdapter{toolID: "tool-b"})

	all := AllAdapters()
	if len(all) != 2 {
		t.Errorf("AllAdapters() returned %d adapters, want 2", len(all))
	}
	if all["tool-a"] == nil {
		t.Error("expected tool-a in AllAdapters")
	}
	if all["tool-b"] == nil {
		t.Error("expected tool-b in AllAdapters")
	}
}
