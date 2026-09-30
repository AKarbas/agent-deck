package session

import (
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/tmux"
	"github.com/stretchr/testify/require"
)

func TestDaemonPreservesPersistedStopAfterLiveSample(t *testing.T) {
	skipIfNoTmuxBinary(t)
	inboxTestHome(t)
	t.Setenv("TMUX", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	const profile = "_test-daemon-stop-review"
	storage, err := NewStorageWithProfile(profile)
	require.NoError(t, err)
	t.Cleanup(func() { _ = storage.Close() })

	inst := NewInstanceWithTool("daemon-stop-review", t.TempDir(), "shell")
	inst.CreatedAt = time.Now().Add(-time.Minute)
	inst.TmuxSocketName = "daemon-stop-review-" + inst.ID
	inst.tmuxSession.SocketName = inst.TmuxSocketName
	require.NoError(t, inst.tmuxSession.Start(""))
	t.Cleanup(func() { _ = inst.Kill() })
	require.NoError(t, storage.SaveWithGroups([]*Instance{inst}, nil))

	d := NewTransitionDaemon()
	t.Cleanup(d.Flush)
	d.storages[profile] = storage
	d.syncProfile(profile)
	prior, ok := d.livePrior[profile][inst.ID]
	require.True(t, ok, "the daemon must first sample the real live shell")
	require.True(t, isLiveSessionStatus(prior.status), "live sample was %q", prior.status)

	// Model a separate CLI stop between daemon polls. The daemon retains
	// its prior, while the freshly loaded row now records the user's stop.
	require.NoError(t, inst.Kill())
	require.False(t, tmux.HasSessionOnSocket(inst.TmuxSocketName, inst.tmuxSession.Name))
	require.NoError(t, storage.SaveWithGroups([]*Instance{inst}, nil))
	for pass := 0; pass < 3; pass++ {
		d.syncProfile(profile)
		require.Equal(t, string(StatusStopped), d.lastStatus[profile][inst.ID])
		loaded, err := storage.Load()
		require.NoError(t, err)
		require.Len(t, loaded, 1)
		require.Equal(t, StatusStopped, loaded[0].Status, "persisted verdict after pass %d", pass)
	}
}
