package cmd

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readTarEntries extracts all entries from a tar archive, returning a map of
// path -> content (empty string for directories).
func readTarEntries(t *testing.T, data []byte) map[string]string {
	t.Helper()
	entries := make(map[string]string)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Typeflag == tar.TypeDir {
			entries[hdr.Name] = ""
			continue
		}
		buf, err := io.ReadAll(tr)
		require.NoError(t, err)
		entries[hdr.Name] = string(buf)
	}
	return entries
}

func TestTarYAMLDir_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "manifests")
	require.NoError(t, os.Mkdir(yamlDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(yamlDir, "app.yaml"), []byte("apiVersion: v1"), 0644))

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	entries := readTarEntries(t, data)

	assert.Contains(t, entries, "manifests/app.yaml")
	assert.Equal(t, "apiVersion: v1", entries["manifests/app.yaml"])
}

func TestTarYAMLDir_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "release")
	subDir := filepath.Join(yamlDir, "charts", "mychart")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(yamlDir, "config.yaml"), []byte("top-level"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "values.yaml"), []byte("nested"), 0644))

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	entries := readTarEntries(t, data)

	assert.Equal(t, "top-level", entries["release/config.yaml"])
	assert.Equal(t, "nested", entries["release/charts/mychart/values.yaml"])
}

func TestTarYAMLDir_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "yamls")
	require.NoError(t, os.Mkdir(yamlDir, 0755))

	files := map[string]string{
		"a.yaml": "a-content",
		"b.yaml": "b-content",
		"c.yaml": "c-content",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(yamlDir, name), []byte(content), 0644))
	}

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	entries := readTarEntries(t, data)

	for name, content := range files {
		key := filepath.Join("yamls", name)
		assert.Equal(t, content, entries[key], "mismatch for %s", name)
	}
}

func TestTarYAMLDir_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.Mkdir(yamlDir, 0755))

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	// Should contain only the top-level directory entry
	entries := readTarEntries(t, data)
	assert.Len(t, entries, 1)
	assert.Contains(t, entries, "empty")
}

func TestTarYAMLDir_PreservesTopLevelDirName(t *testing.T) {
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "my-release-v1.2.3")
	require.NoError(t, os.Mkdir(yamlDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(yamlDir, "test.yaml"), []byte("data"), 0644))

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	entries := readTarEntries(t, data)
	assert.Contains(t, entries, "my-release-v1.2.3/test.yaml")
}

func TestTarYAMLDir_NonExistentDir(t *testing.T) {
	_, err := tarYAMLDir("/nonexistent/path")
	assert.Error(t, err)
}

func TestTarYAMLDir_ProducesValidTar(t *testing.T) {
	// Verify the tar can be fully read without errors and contains the expected
	// number of entries (directories + files).
	tmpDir := t.TempDir()
	yamlDir := filepath.Join(tmpDir, "app")
	sub := filepath.Join(yamlDir, "sub")
	require.NoError(t, os.MkdirAll(sub, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(yamlDir, "root.yaml"), []byte("r"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "child.yaml"), []byte("c"), 0644))

	data, err := tarYAMLDir(yamlDir)
	require.NoError(t, err)

	var names []string
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		names = append(names, hdr.Name)
	}

	sort.Strings(names)
	expected := []string{"app", "app/root.yaml", "app/sub", "app/sub/child.yaml"}
	sort.Strings(expected)
	assert.Equal(t, expected, names)
}
