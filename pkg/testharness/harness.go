package testharness

import (
	"testing"
	"time"

	"github.com/aserto-dev/go-utils/testutil"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/go-sample-project/pkg/app"
	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
)

// TestHarness wraps a GoSampleProject so we can set it up easily
// and monitor its logs
type TestHarness struct {
	GoSampleProject       *app.GoSampleProject
	LogDebugger *testutil.LogDebugger

	cleanup      func()
	t            *testing.T
}

// Cleanup cleans up the application, releasing all resources
func (h *TestHarness) Cleanup() {
	assert := require.New(h.t)
	assert.NoError(h.GoSampleProject.Server.Stop())

	// Cleanup the app
	h.cleanup()

	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8484")
	}, 10*time.Second, 10*time.Millisecond)
	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8383")
	}, 10*time.Second, 10*time.Millisecond)
	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8282")
	}, 10*time.Second, 10*time.Millisecond)
}

// Setup creates a new TestHarness
func Setup(t *testing.T, configOverrides func(*config.Config)) *TestHarness {
	assert := require.New(t)

	var err error
	h := &TestHarness{t: t, LogDebugger: testutil.NewLogDebugger(t, "go-sample-project")}

	h.GoSampleProject, h.cleanup, err = app.BuildTestGoSampleProject(
		h.LogDebugger, h.LogDebugger, AssetDefaultConfig(), configOverrides)
	assert.NoError(err)

	err = h.GoSampleProject.Server.Start()
	assert.NoError(err)

	assert.Eventually(func() bool {
		return testutil.PortOpen("127.0.0.1:8383")
	}, 10*time.Second, 10*time.Millisecond)

	return h
}
