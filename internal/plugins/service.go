package plugins

// Service provides plugin lifecycle operations.
type Service struct {
	Dir string
}

// NewService creates a plugin service rooted at the given plugins directory.
func NewService(dir string) *Service {
	return &Service{Dir: dir}
}

// Discover returns all installed plugins.
func (s *Service) Discover() ([]Plugin, error) {
	return discover(s.Dir)
}

// FindForCommand finds a plugin that handles the given command.
func (s *Service) FindForCommand(cmd string) (Plugin, bool) {
	pluginList, err := s.Discover()
	if err != nil {
		return Plugin{}, false
	}
	return findForCommand(pluginList, cmd)
}
