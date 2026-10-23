package apply

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konveyor/crane/internal/file"
	internalTransform "github.com/konveyor/crane/internal/transform"
	"github.com/sirupsen/logrus"
)

const crdReconciliationResources = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  names:
    kind: Widget
    plural: widgets
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: gadgets.example.com
spec:
  group: example.com
  names:
    kind: Gadget
    plural: gadgets
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: policies.example.com
spec:
  group: example.com
  names:
    kind: Policy
    plural: policies
---
apiVersion: example.com/v2
kind: Gadget
metadata:
  name: retained-gadget
  namespace: source
---
apiVersion: example.com/v1
kind: Policy
metadata:
  name: retained-policy
  namespace: source
`

func TestFilterUnneededAutomaticallyCollectedCRDs(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.ErrorLevel)
	applier := &KustomizeApplier{Log: logger}

	filtered, err := applier.filterUnneededAutomaticallyCollectedCRDs(
		[]byte(crdReconciliationResources),
		[]string{"widgets.example.com", "gadgets.example.com"},
	)
	if err != nil {
		t.Fatalf("filterUnneededAutomaticallyCollectedCRDs: %v", err)
	}
	got := string(filtered)
	if strings.Contains(got, "widgets.example.com") {
		t.Error("unused automatically collected Widget CRD was retained")
	}
	if !strings.Contains(got, "gadgets.example.com") {
		t.Error("Gadget CRD was removed despite a v2 Gadget resource")
	}
	if !strings.Contains(got, "policies.example.com") {
		t.Error("CRD not marked as automatically collected was removed")
	}
}

func TestApplyMultiStageReconcilesAutomaticallyCollectedCRDs(t *testing.T) {
	tests := []struct {
		name          string
		writeMetadata bool
		wantWidgetCRD bool
	}{
		{
			name:          "automatically collected CRD without instances is excluded",
			writeMetadata: true,
			wantWidgetCRD: false,
		},
		{
			name:          "legacy stage without metadata retains CRD",
			writeMetadata: false,
			wantWidgetCRD: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stageDir := filepath.Join(root, "10_KubernetesPlugin")
			outputDir := filepath.Join(root, "output")
			if err := os.MkdirAll(stageDir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(stageDir, "resources.yaml"), []byte(crdReconciliationResources), 0600); err != nil {
				t.Fatal(err)
			}
			kustomization := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n- resources.yaml\n"
			if err := os.WriteFile(filepath.Join(stageDir, "kustomization.yaml"), []byte(kustomization), 0600); err != nil {
				t.Fatal(err)
			}
			if tt.writeMetadata {
				if err := file.WriteMetadata(filepath.Join(stageDir, file.MetadataFilename), &file.Metadata{
					AutomaticallyCollectedCRDs: []string{"widgets.example.com", "gadgets.example.com"},
				}); err != nil {
					t.Fatal(err)
				}
			}

			logger := logrus.New()
			logger.SetLevel(logrus.ErrorLevel)
			applier := &KustomizeApplier{Log: logger, TransformDir: root, OutputDir: outputDir}
			if err := applier.ApplyMultiStage(internalTransform.StageSelector{}); err != nil {
				t.Fatalf("ApplyMultiStage: %v", err)
			}

			output, err := os.ReadFile(filepath.Join(outputDir, "output.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			widgetCRDPath := filepath.Join(outputDir, "resources", "_cluster", "CustomResourceDefinition_apiextensions.k8s.io_v1_clusterscoped_widgets.example.com.yaml")
			_, widgetCRDErr := os.Stat(widgetCRDPath)
			if strings.Contains(string(output), "widgets.example.com") != tt.wantWidgetCRD {
				t.Errorf("output.yaml Widget CRD presence = %t, want %t", strings.Contains(string(output), "widgets.example.com"), tt.wantWidgetCRD)
			}
			if (widgetCRDErr == nil) != tt.wantWidgetCRD {
				t.Errorf("split Widget CRD presence = %t, want %t (error: %v)", widgetCRDErr == nil, tt.wantWidgetCRD, widgetCRDErr)
			}
			if !strings.Contains(string(output), "gadgets.example.com") {
				t.Error("output.yaml omitted CRD required by retained v2 Gadget")
			}
		})
	}
}
