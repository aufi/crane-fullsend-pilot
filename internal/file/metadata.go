package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// MetadataFilename is the sidecar containing Crane-specific export provenance.
// It is not a Kubernetes resource and is ignored when loading manifests.
const MetadataFilename = ".crane-metadata.json"

// Metadata records export provenance needed by later pipeline steps.
type Metadata struct {
	AutomaticallyCollectedCRDs []string `json:"automaticallyCollectedCRDs,omitempty"`
}

// ReadMetadata returns nil when the sidecar is absent so exports produced by
// older Crane versions retain their existing behavior.
func ReadMetadata(path string) (*Metadata, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read metadata file %q: %w", path, err)
	}

	metadata := &Metadata{}
	if err := json.Unmarshal(contents, metadata); err != nil {
		return nil, fmt.Errorf("parse metadata file %q: %w", path, err)
	}
	metadata.AutomaticallyCollectedCRDs = uniqueStrings(metadata.AutomaticallyCollectedCRDs)
	return metadata, nil
}

// WriteMetadata writes export provenance without modifying exported resources.
func WriteMetadata(path string, metadata *Metadata) error {
	if metadata == nil {
		return nil
	}
	metadata.AutomaticallyCollectedCRDs = uniqueStrings(metadata.AutomaticallyCollectedCRDs)
	contents, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata file %q: %w", path, err)
	}
	if err := os.WriteFile(path, append(contents, '\n'), 0600); err != nil {
		return fmt.Errorf("write metadata file %q: %w", path, err)
	}
	return nil
}

func uniqueStrings(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
