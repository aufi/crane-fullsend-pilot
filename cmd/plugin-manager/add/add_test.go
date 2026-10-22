package add

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDownloadBinaryUsesSafePermissions(t *testing.T) {
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("plugin"), 0600); err != nil {
		t.Fatalf("failed to create source plugin: %v", err)
	}

	previousUmask := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(previousUmask) })

	pluginDir := filepath.Join(tempDir, "plugins", "nested")
	if err := downloadBinary(pluginDir, "example", source, logrus.New()); err != nil {
		t.Fatalf("downloadBinary() error = %v", err)
	}

	dirInfo, err := os.Stat(pluginDir)
	if err != nil {
		t.Fatalf("failed to stat plugin directory: %v", err)
	}
	if got, want := dirInfo.Mode().Perm(), os.FileMode(0755); got != want {
		t.Errorf("plugin directory mode = %04o, want %04o", got, want)
	}

	fileInfo, err := os.Stat(filepath.Join(pluginDir, "example"))
	if err != nil {
		t.Fatalf("failed to stat plugin binary: %v", err)
	}
	if got, want := fileInfo.Mode().Perm(), os.FileMode(0755); got != want {
		t.Errorf("plugin binary mode = %04o, want %04o", got, want)
	}
}

func TestDownloadBinarySecuresExistingPluginDirectory(t *testing.T) {
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("plugin"), 0600); err != nil {
		t.Fatalf("failed to create source plugin: %v", err)
	}

	previousUmask := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(previousUmask) })

	pluginDir := filepath.Join(tempDir, "plugins")
	if err := os.Mkdir(pluginDir, 0777); err != nil {
		t.Fatalf("failed to create plugin directory: %v", err)
	}

	if err := downloadBinary(pluginDir, "example", source, logrus.New()); err != nil {
		t.Fatalf("downloadBinary() error = %v", err)
	}

	dirInfo, err := os.Stat(pluginDir)
	if err != nil {
		t.Fatalf("failed to stat plugin directory: %v", err)
	}
	if got, want := dirInfo.Mode().Perm(), os.FileMode(0755); got != want {
		t.Errorf("plugin directory mode = %04o, want %04o", got, want)
	}
}

func TestDownloadBinaryRejectsInsecurePluginDirectoryAncestor(t *testing.T) {
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("plugin"), 0600); err != nil {
		t.Fatalf("failed to create source plugin: %v", err)
	}

	previousUmask := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(previousUmask) })

	parentDir := filepath.Join(tempDir, "plugins")
	if err := os.Mkdir(parentDir, 0777); err != nil {
		t.Fatalf("failed to create plugin directory parent: %v", err)
	}
	pluginDir := filepath.Join(parentDir, "nested")

	if err := downloadBinary(pluginDir, "example", source, logrus.New()); err == nil {
		t.Fatal("downloadBinary() succeeded with a group/other-writable plugin directory ancestor")
	}
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Fatalf("plugin directory exists after rejected install, stat error = %v", err)
	}
}
