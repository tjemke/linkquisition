package linkquisition

// Paths holds the folders linkquisition keeps its own files in. Each platform
// resolves them once at startup, and callers receive the whole set instead of
// one parameter or getter per folder.
type Paths struct {
	// Config is the folder holding config.json.
	Config string

	// Log is the folder holding linkquisition.log.
	Log string

	// Plugins is the folder relative plugin paths are resolved against.
	// It is empty on platforms without plugin support.
	Plugins string
}
