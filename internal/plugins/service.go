package plugins

import "sync"

// Service provides plugin lifecycle operations.
type Service struct {
	Dir string

	discoverOnce sync.Once
	discovered   []Plugin
	discoverErr  error
}

// NewService creates a plugin service rooted at the given plugins directory.
func NewService(dir string) *Service {
	return &Service{Dir: dir}
}

// Discover returns all installed plugins. The underlying filesystem scan
// runs at most once per Service instance; subsequent calls return the
// cached result (including the cached error, if the first scan failed).
func (s *Service) Discover() ([]Plugin, error) {
	s.discoverOnce.Do(func() {
		s.discovered, s.discoverErr = discover(s.Dir)
	})
	return s.discovered, s.discoverErr
}

// FindForCommand finds a plugin that handles the given command.
func (s *Service) FindForCommand(cmd string) (Plugin, bool) {
	pluginList, err := s.Discover()
	if err != nil {
		return Plugin{}, false
	}
	return findForCommand(pluginList, cmd)
}
