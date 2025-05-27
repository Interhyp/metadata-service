package check

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/stretchr/testify/assert"
	"io"
	"reflect"
	"testing"
)

func TestMetadataYamlFileWalker_fixExemptionsInFile(t *testing.T) {
	type args struct {
		path         string
		fileContents []byte
	}
	type want struct {
		result any
	}
	type mock struct {
		config *Config
	}
	hasMock := func(m mock) bool {
		return !reflect.DeepEqual(m, mock{config: &Config{}})
	}
	tests := []struct {
		name string
		args args
		want want
		mock mock
	}{
		{
			name: "not in /repository",
			args: args{
				path:         "some/other/path.yaml",
				fileContents: []byte("attribute: value\notherAttribute: otherValue"),
			},
			want: want{
				result: nil,
			},
		},
		{
			name: "repositoryType not matching",
			mock: mock{
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"existing-exemption-one", "existing-exemption-two"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.other-repository-type.yaml",
				fileContents: []byte(`url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: some-ref-matcher
      exemptions:
        - existing-exemption-one
        - existing-exemption-two
`),
			},
			want: want{
				result: nil,
			},
		},
		{
			name: "no missing exemptions",
			mock: mock{
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"existing-exemption-one", "existing-exemption-two"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				fileContents: []byte(`url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: some-ref-matcher
      exemptions:
        - existing-exemption-one
        - existing-exemption-two
`),
			},
			want: want{
				result: nil,
			},
		},
		{
			name: "refMatcher not matching",
			mock: mock{
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"existing-exemption-one", "existing-exemption-two"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				fileContents: []byte(`url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: other-ref-matcher
      exemptions:
        - existing-exemption-one
        - existing-exemption-two
`),
			},
			want: want{
				result: nil,
			},
		},
		{
			name: "fix exemptions in requireConditions",
			mock: mock{
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"existing-exemption", "missing-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				fileContents: []byte(`url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: some-ref-matcher
      exemptions:
        - existing-exemption
`),
			},
			want: want{
				result: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: some-ref-matcher
      exemptions:
        - existing-exemption
        - missing-exemption
`,
			},
		},
		{
			name: "fix exemptions in refProtections",
			mock: mock{
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "tags.preventCreation",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"existing-exemption", "missing-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				fileContents: []byte(`url: existing-repo-url
mainline: master
configuration:
  refProtections:
    tags:
      preventCreation:
        - pattern: some-ref-matcher
          exemptions:
            - existing-exemption
`),
			},
			want: want{
				result: `url: existing-repo-url
mainline: master
configuration:
  refProtections:
    tags:
      preventCreation:
        - pattern: some-ref-matcher
          exemptions:
            - existing-exemption
            - missing-exemption
`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MetadataYamlFileWalker(nil)
			if hasMock(tt.mock) {
				v.fs = memfs.New()
				_, _ = v.fs.Create(tt.args.path)
				if tt.mock.config != nil {
					v.config = *tt.mock.config
				}
			}
			if got := v.fixExemptionsInFile(tt.args.fileContents, tt.args.path); got == nil {
				file, err := v.fs.Open(tt.args.path)
				assert.NoError(t, err)
				defer file.Close()
				content, err := io.ReadAll(file)
				assert.NoError(t, err)
				if tt.want.result != nil {
					assert.Equal(t, tt.want.result, string(content))
				} else {
					assert.Equal(t, "", string(content))
				}
			}
		})
	}
}
