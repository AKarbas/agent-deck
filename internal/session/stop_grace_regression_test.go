package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStoppedSessionSurvivesStartupGraceRefresh(t *testing.T) {
	skipIfNoTmuxBinary(t)

	inst := NewInstanceWithTool("stop-grace", t.TempDir(), "shell")
	inst.Command = "sleep 60"
	require.NoError(t, inst.Start())
	t.Cleanup(func() { _ = inst.Kill() })
	require.NoError(t, inst.Kill())
	require.False(t, inst.Exists(), "stop must remove the tmux session")

	start := time.Now()
	for _, at := range []time.Duration{0, 500 * time.Millisecond, 5 * time.Second} {
		if wait := at - time.Since(start); wait > 0 {
			time.Sleep(wait)
		}
		inst.ForceNextStatusCheck()
		require.NoError(t, inst.UpdateStatus())
		require.Equalf(t, StatusStopped, inst.GetStatusThreadSafe(), "status at +%s", at)
		require.Falsef(t, inst.Exists(), "tmux session at +%s", at)
	}
}
