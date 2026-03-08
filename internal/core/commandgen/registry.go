package commandgen

var registry = map[string]ToolCommandAdapter{}

// Register adds an adapter to the global registry.
func Register(adapter ToolCommandAdapter) {
	registry[adapter.GetToolID()] = adapter
}

// Get returns the adapter for the given tool ID, or nil if not found.
func Get(toolID string) ToolCommandAdapter {
	return registry[toolID]
}

// AllAdapters returns all registered adapters.
func AllAdapters() map[string]ToolCommandAdapter {
	return registry
}
