package engine

import (
	"context"
	"os"

	"github.com/monkeymonk/gdt/internal/metadata"
	"github.com/monkeymonk/gdt/internal/platform"
)

// Install downloads and installs a Godot engine version.
func (s *Service) Install(ctx context.Context, version string, opts InstallOpts) (*InstallResult, error) {
	return s.downloadAndInstall(ctx, downloadSpec{
		CachePath: s.CachePath(),
		APIURL:    s.Config.GodotAPIURL(),
		Token:     os.Getenv("GITHUB_TOKEN"),
		Query:     version,
		Mono:      opts.Mono,
		Force:     opts.Force,
		Refresh:   opts.Refresh,
		DestDir:   s.VersionsDir(),
		ResolveArtifact: func(release *metadata.Release, plat platform.Info, mono bool) (string, error) {
			return metadata.ResolveEngineArtifact(release, plat, mono)
		},
		PostInstall:    s.installDesktop,
		VerifyChecksum: true,
	})
}
