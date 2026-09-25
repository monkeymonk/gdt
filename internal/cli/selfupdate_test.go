package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
)

// TestSelfUpdateCmd_AlreadyUpToDate proves that when selfupdate.Update
// reports no available update (current version matches the latest release
// tag), "gdt self update" prints the "already up to date" message rather
// than the "updated gdt to %s" message.
//
// selfupdate.Update's short-circuit for the already-up-to-date case returns
// before any binary download/replace step, so it is safely reachable from
// this package via the same apiURL injection app.Config.SelfUpdateAPIURL()
// already provides (mirroring the GodotAPIURL seam used by install.go and
// update.go). The Updated == true branch is NOT covered here: driving it
// requires selfupdate.Update to actually replace the file at
// os.Executable(), and the only seam for faking that path (the unexported
// osExecutable package var) is only reachable from within the selfupdate
// package's own tests, not from cli. Exercising the real branch from here
// would mean letting Update() overwrite this test binary in place, which is
// unsafe and non-deterministic, so that branch is accepted-incomplete.
func TestSelfUpdateCmd_AlreadyUpToDate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"tag_name": "v1.0.0"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	app := &App{
		Version: "1.0.0",
		Config:  &config.Config{SelfUpdateAPI: srv.URL + "/release"},
	}

	cmd := newSelfUpdateCmd(app)
	cmd.SetArgs([]string{"update"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if !strings.Contains(output, "already up to date") {
		t.Errorf("expected \"already up to date\" message, got: %q", output)
	}
	if strings.Contains(output, "updated gdt to") {
		t.Errorf("did not expect the \"updated gdt to\" message, got: %q", output)
	}
}
