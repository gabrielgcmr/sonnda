// internal/features/documentprocessing/extraction/snapshot.go
package extraction

import (
	"encoding/json"
	"fmt"

	"github.com/gabrielgcmr/sonnda/internal/domain/labextraction"
)

// Snapshot keeps fields deliberately excluded from the public extraction JSON.
type Snapshot struct {
	Version  int                              `json:"version"`
	Result   Result                           `json:"result"`
	Metadata labextraction.ExtractionMetadata `json:"metadata"`
	RawText  *string                          `json:"raw_text"`
	Tests    []TestMetadata                   `json:"tests_metadata"`
}
type NodeMetadata struct {
	RawText    *string                           `json:"raw_text"`
	Status     labextraction.ExtractionStatus    `json:"status"`
	Confidence *float64                          `json:"confidence"`
	Warnings   []labextraction.ExtractionWarning `json:"warnings"`
}
type TestMetadata struct {
	NodeMetadata
	Items []NodeMetadata `json:"items"`
}

func Encode(result *Result) ([]byte, error) {
	s := Snapshot{Version: 1, Result: *result, Metadata: result.Report.Metadata, RawText: result.Report.RawText}
	for _, test := range result.Report.Tests {
		metadata := TestMetadata{NodeMetadata: NodeMetadata{test.RawText, test.Status, test.Confidence, test.Warnings}}
		for _, item := range test.Items {
			metadata.Items = append(metadata.Items, NodeMetadata{item.RawText, item.Status, item.Confidence, item.Warnings})
		}
		s.Tests = append(s.Tests, metadata)
	}
	return json.Marshal(s)
}

func Decode(data []byte) (*Result, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Version != 1 {
		return nil, fmt.Errorf("unsupported extraction snapshot version %d", s.Version)
	}
	if len(s.Tests) != len(s.Result.Report.Tests) {
		return nil, fmt.Errorf("invalid extraction snapshot tests")
	}
	s.Result.Report.Metadata, s.Result.Report.RawText = s.Metadata, s.RawText
	for i, meta := range s.Tests {
		test := &s.Result.Report.Tests[i]
		test.RawText, test.Status, test.Confidence, test.Warnings = meta.RawText, meta.Status, meta.Confidence, meta.Warnings
		if len(meta.Items) != len(test.Items) {
			return nil, fmt.Errorf("invalid extraction snapshot items")
		}
		for j, m := range meta.Items {
			item := &test.Items[j]
			item.RawText, item.Status, item.Confidence, item.Warnings = m.RawText, m.Status, m.Confidence, m.Warnings
		}
	}
	return &s.Result, nil
}
