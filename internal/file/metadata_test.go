package file

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMetadataIsIgnoredWhenReadingResources(t *testing.T) {
	dir := t.TempDir()
	if err := WriteMetadata(filepath.Join(dir, MetadataFilename), &Metadata{
		AutomaticallyCollectedCRDs: []string{"widgets.example.com"},
	}); err != nil {
		t.Fatal(err)
	}
	resource := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n"
	if err := os.WriteFile(filepath.Join(dir, "configmap.yaml"), []byte(resource), 0600); err != nil {
		t.Fatal(err)
	}

	resources, err := ReadFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("ReadFiles: %v", err)
	}
	if len(resources) != 1 || resources[0].Unstructured.GetName() != "test" {
		t.Fatalf("resources = %#v, want only ConfigMap/test", resources)
	}
	metadata, err := ReadMetadata(filepath.Join(dir, MetadataFilename))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(metadata.AutomaticallyCollectedCRDs, []string{"widgets.example.com"}) {
		t.Errorf("AutomaticallyCollectedCRDs = %v", metadata.AutomaticallyCollectedCRDs)
	}
}
