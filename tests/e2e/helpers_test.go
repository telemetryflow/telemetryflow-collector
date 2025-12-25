package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	projectRoot     string
	projectRootOnce sync.Once
	binaryPath      string
	buildOnce       sync.Once
	buildErr        error
)

// getProjectRoot returns the absolute path to the project root directory.
func getProjectRoot(t *testing.T) string {
	t.Helper()
	projectRootOnce.Do(func() {
		dir, err := os.Getwd()
		require.NoError(t, err)

		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				projectRoot = dir
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				require.Fail(t, "could not find project root")
			}
			dir = parent
		}
	})
	return projectRoot
}

// getCollectorBinary returns the path to the collector binary.
// It uses TFO_COLLECTOR_BINARY env var if set, otherwise builds the binary.
func getCollectorBinary(t *testing.T) string {
	t.Helper()

	// Check if binary path is provided via environment variable
	if envBinary := os.Getenv("TFO_COLLECTOR_BINARY"); envBinary != "" {
		// Make path absolute if relative
		if !filepath.IsAbs(envBinary) {
			envBinary = filepath.Join(getProjectRoot(t), envBinary)
		}
		if _, err := os.Stat(envBinary); err == nil {
			return envBinary
		}
	}

	root := getProjectRoot(t)
	buildOnce.Do(func() {
		binaryPath = filepath.Join(root, "build", "tfo-collector-e2e-test")
		_ = os.MkdirAll(filepath.Join(root, "build"), 0755)

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/tfo-collector")
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = &buildError{err: err, output: string(output)}
		}
	})

	require.NoError(t, buildErr, "failed to build collector")
	return binaryPath
}

// getTestdataPath returns the absolute path to a testdata file.
func getTestdataPath(t *testing.T, filename string) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "tests", "e2e", "testdata", filename)
}

type buildError struct {
	err    error
	output string
}

func (e *buildError) Error() string {
	return e.err.Error() + "\n" + e.output
}
