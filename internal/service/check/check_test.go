package check

import (
	"fmt"
	"github.com/google/go-github/v70/github"
	"reflect"
	"testing"
)

func Test_walkerToCheckRunOutput(t *testing.T) {
	annotations := []*github.CheckRunAnnotation{
		{
			Path:            github.Ptr("some/path/to/a/file.yaml"),
			StartLine:       github.Ptr(1),
			EndLine:         github.Ptr(1),
			AnnotationLevel: github.Ptr("failure"),
			Message:         github.Ptr("test message"),
			Title:           github.Ptr("test title"),
		},
	}
	type args struct {
		johnnie *MetadataWalker
	}
	tests := []struct {
		name string
		args args
		want github.CheckRunOutput
	}{
		{
			name: "passed",
			args: args{
				johnnie: &MetadataWalker{
					Annotations: make([]*github.CheckRunAnnotation, 0),
					Errors:      make(map[string]error),
				},
			},
			want: github.CheckRunOutput{
				Title:            github.Ptr("Passed YAML validation"),
				Summary:          github.Ptr("All changed files are valid."),
				Text:             nil,
				AnnotationsCount: nil,
				AnnotationsURL:   nil,
				Annotations:      make([]*github.CheckRunAnnotation, 0),
				Images:           nil,
			},
		},
		{
			name: "failed with findings",
			args: args{
				johnnie: &MetadataWalker{
					Annotations: annotations,
					Errors:      make(map[string]error),
				},
			},
			want: github.CheckRunOutput{
				Title:            github.Ptr("Failed YAML validation"),
				Summary:          github.Ptr("There were files failing the validation. See Annotations."),
				Text:             nil,
				AnnotationsCount: nil,
				AnnotationsURL:   nil,
				Annotations:      annotations,
				Images:           nil,
			},
		},
		{
			name: "failed with errors",
			args: args{
				johnnie: &MetadataWalker{
					Annotations: make([]*github.CheckRunAnnotation, 0),
					Errors: map[string]error{
						"some/path/to/a/failed/file.yaml":       fmt.Errorf("first test failure"),
						"some/path/to/another/failed/file.yaml": fmt.Errorf("second test failure"),
					},
				},
			},
			want: github.CheckRunOutput{
				Title:            github.Ptr("Failed YAML validation"),
				Summary:          github.Ptr("There were files causing errors. See Details."),
				Text:             github.Ptr("The following validation errors occurred:\n- some/path/to/a/failed/file.yaml: first test failure\n- some/path/to/another/failed/file.yaml: second test failure\n"),
				AnnotationsCount: nil,
				AnnotationsURL:   nil,
				Annotations:      make([]*github.CheckRunAnnotation, 0),
				Images:           nil,
			},
		},
		{
			name: "failed with findings and errors",
			args: args{
				johnnie: &MetadataWalker{
					Annotations: annotations,
					Errors: map[string]error{
						"some/path/to/a/failed/file.yaml":       fmt.Errorf("first test failure"),
						"some/path/to/another/failed/file.yaml": fmt.Errorf("second test failure"),
					},
				},
			},
			want: github.CheckRunOutput{
				Title:            github.Ptr("Failed YAML validation"),
				Summary:          github.Ptr("There were files failing the validation. See Annotations.\nThere were files causing errors. See Details."),
				Text:             github.Ptr("The following validation errors occurred:\n- some/path/to/a/failed/file.yaml: first test failure\n- some/path/to/another/failed/file.yaml: second test failure\n"),
				AnnotationsCount: nil,
				AnnotationsURL:   nil,
				Annotations:      annotations,
				Images:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := walkerToCheckRunOutput(tt.args.johnnie); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("walkerToCheckRunOutput() = %+v, want %+v", printOutput(got), printOutput(tt.want))
			}
		})
	}
}

func printOutput(in github.CheckRunOutput) string {
	return fmt.Sprintf("{Title: %s, Summary: %s, Text: %s, Annotations: %v}", ptrStr(in.Title), ptrStr(in.Summary), ptrStr(in.Text), in.Annotations)
}

func ptrStr(v *string) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s", *v)
}
func ptrInt(v *int) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *v)
}
