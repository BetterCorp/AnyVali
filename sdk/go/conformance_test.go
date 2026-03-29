package anyvali

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ConformanceTestCase represents a single test case from the spec corpus.
type ConformanceTestCase struct {
	Name     string         `json:"name"`
	Schema   map[string]any `json:"schema"`
	Input    any            `json:"input"`
	Expected struct {
		Success bool   `json:"success"`
		Data    any    `json:"data,omitempty"`
		Issues  []struct {
			Code string `json:"code"`
			Path []any  `json:"path,omitempty"`
		} `json:"issues,omitempty"`
	} `json:"expected"`
}

// ConformanceTestFile represents a file of conformance tests.
type ConformanceTestFile struct {
	Description string                `json:"description"`
	Tests       []ConformanceTestCase `json:"tests"`
}

func TestConformance(t *testing.T) {
	corpusDir := filepath.Join("..", "..", "spec", "corpus")

	// Walk all JSON files in the corpus directory
	err := filepath.Walk(corpusDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors accessing the path
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("skipping %s: %v", path, err)
			return nil
		}

		var testFile ConformanceTestFile
		if err := json.Unmarshal(data, &testFile); err != nil {
			t.Logf("skipping %s: failed to parse: %v", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(corpusDir, path)

		for _, tc := range testFile.Tests {
			testName := relPath + "/" + tc.Name
			t.Run(testName, func(t *testing.T) {
				// Build the document
				doc := &Document{
					AnyvaliVersion: "1.0",
					SchemaVersion:  "1",
					Root:           tc.Schema,
				}

				schema, err := Import(doc)
				if err != nil {
					t.Fatalf("failed to import schema: %v", err)
				}

				result := schema.SafeParse(tc.Input)

				if result.Success != tc.Expected.Success {
					t.Errorf("expected success=%v, got success=%v", tc.Expected.Success, result.Success)
					if !result.Success {
						for _, issue := range result.Issues {
							t.Logf("  issue: [%s] %s", issue.Code, issue.Message)
						}
					}
				}

				if tc.Expected.Success && result.Success {
					// Compare data
					expectedJSON, _ := json.Marshal(tc.Expected.Data)
					actualJSON, _ := json.Marshal(result.Data)
					if string(expectedJSON) != string(actualJSON) {
						t.Errorf("data mismatch:\n  expected: %s\n  actual:   %s", expectedJSON, actualJSON)
					}
				}

				if !tc.Expected.Success && !result.Success {
					// Check issue codes
					for _, expectedIssue := range tc.Expected.Issues {
						found := false
						for _, actualIssue := range result.Issues {
							if actualIssue.Code == expectedIssue.Code {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("expected issue code %q not found in result", expectedIssue.Code)
						}
					}
				}
			})
		}
		return nil
	})

	if err != nil {
		t.Logf("could not walk corpus directory: %v (this is OK if corpus is empty)", err)
	}
}
