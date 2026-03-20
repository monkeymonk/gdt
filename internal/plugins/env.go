package plugins

// EnvContext holds the values needed to construct plugin environment variables.
type EnvContext struct {
	Home         string
	ProjectRoot  string
	GodotVersion string
	EnginePath   string
}

// BuildEnv constructs the environment variable slice for plugin subprocesses.
func BuildEnv(ctx EnvContext) []string {
	return []string{
		"GDT_HOME=" + ctx.Home,
		"GDT_PROJECT_ROOT=" + ctx.ProjectRoot,
		"GDT_GODOT_VERSION=" + ctx.GodotVersion,
		"GDT_ENGINE_PATH=" + ctx.EnginePath,
	}
}
