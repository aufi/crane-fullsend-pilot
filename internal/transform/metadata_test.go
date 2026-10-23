package transform

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/konveyor/crane/internal/file"
	"github.com/sirupsen/logrus"
)

func TestRunMultiStagePropagatesExportMetadata(t *testing.T) {
	root := t.TempDir()
	exportDir := filepath.Join(root, "export")
	transformDir := filepath.Join(root, "transform")
	resourceDir := filepath.Join(exportDir, "resources", "source")
	if err := os.MkdirAll(resourceDir, 0700); err != nil {
		t.Fatal(err)
	}
	resource := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n  namespace: source\n"
	if err := os.WriteFile(filepath.Join(resourceDir, "configmap.yaml"), []byte(resource), 0600); err != nil {
		t.Fatal(err)
	}
	if err := file.WriteMetadata(filepath.Join(exportDir, file.MetadataFilename), &file.Metadata{
		AutomaticallyCollectedCRDs: []string{"widgets.example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(transformDir, "10_ManualStage"), 0700); err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	orchestrator := &Orchestrator{
		Log:          logger,
		ExportDir:    exportDir,
		TransformDir: transformDir,
		PluginDir:    filepath.Join(root, "plugins"),
		Overwrite:    true,
	}
	if err := orchestrator.RunMultiStage(StageSelector{}); err != nil {
		t.Fatalf("RunMultiStage: %v", err)
	}

	metadata, err := file.ReadMetadata(filepath.Join(transformDir, "10_ManualStage", file.MetadataFilename))
	if err != nil {
		t.Fatal(err)
	}
	if metadata == nil || !reflect.DeepEqual(metadata.AutomaticallyCollectedCRDs, []string{"widgets.example.com"}) {
		t.Fatalf("stage metadata = %#v, want widgets.example.com", metadata)
	}
}
