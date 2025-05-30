package check

import (
	"fmt"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/require"
	"reflect"
	"strings"
	"testing"
)

func TestMetadataYamlFileWalker_validateSingleYamlFile(t *testing.T) {
	type want struct {
		result  []*github.CheckRunAnnotation
		ignored map[string]string
		errors  map[string]error
	}
	type args struct {
		path     string
		contents string
	}
	type mock struct {
		walkedRepos walkedRepos
		config      *Config
	}
	hasMock := func(m mock) bool {
		return !reflect.DeepEqual(m, mock{walkedRepos: walkedRepos{}})
	}
	tests := []struct {
		name string
		mock mock
		args args
		want want
	}{
		{
			name: "not in owners/",
			args: args{
				path:     "some/other/path.yaml",
				contents: "attribute: value\notherAttribute: otherValue",
			},
			want: want{
				result: nil,
				ignored: map[string]string{
					"some/other/path.yaml": "file is not a .yaml or not situated in owners/",
				},
				errors: make(map[string]error),
			},
		},
		{
			name: "not a yaml",
			args: args{
				path:     "owners/some-owner/owner.info.txt",
				contents: "anything but yaml",
			},
			want: want{
				result: nil,
				ignored: map[string]string{
					"owners/some-owner/owner.info.txt": "file is not a .yaml or not situated in owners/",
				},
				errors: make(map[string]error),
			},
		},
		{
			name: "not a validated file",
			args: args{
				path:     "owners/some-owner/any-other.yaml",
				contents: "attribute: value\notherAttribute: otherValue",
			},
			want: want{
				result: nil,
				ignored: map[string]string{
					"owners/some-owner/any-other.yaml": "file is neither owner info, nor service nor repository",
				},
				errors: make(map[string]error),
			},
		},
		{
			name: "valid owner",
			args: args{
				path: "owners/some-owner/owner.info.yaml",
				contents: `contact: some@mail.com
teamsChannelURL: https://teams.microsoft.com/someChannel
productOwner: someone
groups:
  users:
    - userA
displayName: Some Name
members:
  - userB
`,
			},
			want: want{
				result:  nil,
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid owner - invalid yaml",
			args: args{
				path: "owners/some-owner/owner.info.yaml",
				contents: `contact: some@mail.com
teamsChannelURL: https://teams.microsoft.com/someChannel
productOwner: someone
groups
  users:
    - userA
displayName: Some Name
members:
  - userB
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/owner.info.yaml"),
						StartLine:       github.Ptr(4),
						EndLine:         github.Ptr(4),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("could not find expected ':'"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid owner - invalid field",
			args: args{
				path: "owners/some-owner/owner.info.yaml",
				contents: `contact: some@mail.com
teamsChannelURL: https://teams.microsoft.com/someChannel
productOwner: someone
group:
  users:
    - userA
displayName: Some Name
members:
  - userB
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/owner.info.yaml"),
						StartLine:       github.Ptr(4),
						EndLine:         github.Ptr(4),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("field group not found in type openapi.OwnerDto"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "valid service",
			args: args{
				path: "owners/some-owner/services/service.yaml",
				contents: `description: test
quicklinks: []
repositories: []
alertTarget: some@mail.com
internetExposed: true
lifecycle: experimental
`,
			},
			want: want{
				result:  nil,
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid service - invalid yaml",
			args: args{
				path: "owners/some-owner/services/service.yaml",
				contents: `description: test
quicklinks: [ ]
repositories: [ ]
alertTarget some@mail.com
internetExposed: true
lifecycle: experimental`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/services/service.yaml"),
						StartLine:       github.Ptr(4),
						EndLine:         github.Ptr(4),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("could not find expected ':'"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid service - invalid field",
			args: args{
				path: "owners/some-owner/services/service.yaml",
				contents: `description: test
quicklinks: []
repositories: []
alertTargets: some@mail.com
internetExposed: true
lifecycle: experimental
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/services/service.yaml"),
						StartLine:       github.Ptr(4),
						EndLine:         github.Ptr(4),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("field alertTargets not found in type openapi.ServiceDto"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "valid repo",
			args: args{
				path: "owners/some-owner/repositories/repository.none.yaml",
				contents: `url: ssh://git@server.com/owner/repo.git
mainline: master
configuration:
  requireSuccessfulBuilds: 2
`,
			},
			want: want{
				result:  nil,
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - invalid yaml",
			args: args{
				path: "owners/some-owner/repositories/repository.none.yaml",
				contents: `url: ssh://git@server.com/owner/repo.git
mainline master
configuration:
  requireSuccessfulBuilds: 2`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.none.yaml"),
						StartLine:       github.Ptr(2),
						EndLine:         github.Ptr(2),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("could not find expected ':'"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - invalid field",
			args: args{
				path: "owners/some-owner/repositories/repository.none.yaml",
				contents: `url: ssh://git@server.com/owner/repo.git
mainlines: master
configuration:
  requireSuccessfulBuilds: 2
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.none.yaml"),
						StartLine:       github.Ptr(2),
						EndLine:         github.Ptr(2),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("field mainlines not found in type openapi.RepositoryDto"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - duplicate url",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: map[string]string{
						"existing-repo-url": "owners/some-owner/repositories/other-repository.none.yaml",
					},
					keyToPath: make(map[string]string),
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.none.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireSuccessfulBuilds: 2
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.none.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("Repository url already used by owners/some-owner/repositories/other-repository.none.yaml"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - duplicate key",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{
						"repository.none": "owners/some-other-owner/repositories/repository.none.yaml",
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.none.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireSuccessfulBuilds: 2
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.none.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("failure"),
						Message:         github.Ptr("Repository key already used by owners/some-other-owner/repositories/repository.none.yaml"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - missing condition in requireConditions",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "missing-exemption-two"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    other-condition:
      exemptions:
        - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection some-condition with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - ref not matching in requireConditions",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "missing-exemption-two"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: "other-ref-matcher"
      exemptions:
        - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection some-condition with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - missing exemption in requireConditions",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "some-condition",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-condition:
      refMatcher: "some-ref-matcher"
      exemptions:
        - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain all required exemptions missing-exemption-one, some-exemption for condition some-condition with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected required exemptions"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - missing condition in refProtections",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "branches.preventDeletion",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  refProtections:
    branches:
      requirePR:
        - exemptions:
            - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection branches.preventDeletion with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - ref not matching in refProtections",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "branches.preventDeletion",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  refProtections:
    branches:
      preventDeletion:
        - pattern: "other-ref-matcher"
          exemptions:
            - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection branches.preventDeletion with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - missing exemption in refProtections",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "branches.preventDeletion",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  refProtections:
    branches:
      preventDeletion:
        - pattern: "some-ref-matcher"
          exemptions:
            - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain all required exemptions missing-exemption-one, some-exemption for condition branches.preventDeletion with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected required exemptions"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "invalid repo - missing exemption in requiredConditions and refProtections",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "branches.preventDeletion",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
						{
							Name:       "some-name",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"missing-exemption-one", "some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    other-condition:
      exemptions:
        - some-exemption
  refProtections:
    branches:
      requirePR:
        - pattern: "some-ref-matcher"
          exemptions:
            - some-exemption
`,
			},
			want: want{
				result: []*github.CheckRunAnnotation{
					{
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection branches.preventDeletion with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					}, {
						Path:            github.Ptr("owners/some-owner/repositories/repository.helm-deployment.yaml"),
						StartLine:       github.Ptr(1),
						EndLine:         github.Ptr(1),
						AnnotationLevel: github.Ptr("warning"),
						Message:         github.Ptr("This file does not contain the required condition/refProtection some-name with the refMatcher some-ref-matcher."),
						Title:           github.Ptr("missing expected condition/refProtection"),
					},
				},
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
		{
			name: "valid repo checkRequiredConditions",
			mock: mock{
				walkedRepos: walkedRepos{
					urlToPath: make(map[string]string),
					keyToPath: map[string]string{},
				},
				config: &Config{
					expectedExemptions: []config.CheckedExpectedExemption{
						{
							Name:       "branches.preventDeletion",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"some-exemption"},
						},
						{
							Name:       "some-name",
							RefMatcher: "some-ref-matcher",
							Exemptions: []string{"some-exemption"},
						},
					},
				},
			},
			args: args{
				path: "owners/some-owner/repositories/repository.helm-deployment.yaml",
				contents: `url: existing-repo-url
mainline: master
configuration:
  requireConditions:
    some-name:
      refMatcher: "some-ref-matcher"
      exemptions:
        - some-exemption
  refProtections:
    branches:
      preventDeletion:
        - pattern: "some-ref-matcher"
          exemptions:
            - some-exemption
`,
			},
			want: want{
				result:  nil,
				ignored: make(map[string]string),
				errors:  make(map[string]error),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := MetadataYamlFileWalker(nil)
			if hasMock(tt.mock) {
				v.walkedRepos = tt.mock.walkedRepos
				if tt.mock.config != nil {
					v.config = *tt.mock.config
				}
			}
			if got := v.validateSingleYamlFile(tt.args.path, tt.args.contents); !reflect.DeepEqual(got, tt.want.result) {
				t.Errorf("validateSingleYamlFile() = %v, want %v", printAnnotations(got), printAnnotations(tt.want.result))
			}
			require.Equal(t, tt.want.ignored, v.IgnoredWithReason)
			require.Equal(t, tt.want.errors, v.Errors)
		})
	}
}

func printAnnotations(in []*github.CheckRunAnnotation) string {
	sb := strings.Builder{}
	sb.WriteRune('[')
	for i, a := range in {
		sb.WriteString(fmt.Sprintf("%d: ", i))
		if a == nil {
			sb.WriteString("<nil>,")
			continue
		}
		sb.WriteRune('{')
		sb.WriteString(fmt.Sprintf("message: \"%s\",", ptrStr(a.Message)))
		sb.WriteString(fmt.Sprintf("path: \"%s\",", ptrStr(a.Path)))
		sb.WriteString(fmt.Sprintf("title: \"%s\",", ptrStr(a.Title)))
		sb.WriteString(fmt.Sprintf("startLine: %s,", ptrInt(a.StartLine)))
		sb.WriteString(fmt.Sprintf("endLine: %s,", ptrInt(a.EndLine)))
		sb.WriteRune('}')
	}
	sb.WriteRune(']')
	return sb.String()
}
