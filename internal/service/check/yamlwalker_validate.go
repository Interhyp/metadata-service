package check

import (
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/go-git/go-billy/v5/util"
	"github.com/google/go-github/v70/github"
	"github.com/google/yamlfmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
)

func (v *MetadataWalker) ValidateMetadata() error {
	return util.Walk(v.fs, v.rootDir, v.validateWalkFunc)
}

func (v *MetadataWalker) validateWalkFunc(path string, info fs.FileInfo, err error) error {
	return v.walkFunc(
		func(fileContents []byte) error {
			trimmed := strings.Trim(path, "/")
			annotations := v.validateSingleYamlFile(trimmed, string(fileContents))
			if len(annotations) > 0 {
				v.Annotations = append(v.Annotations, annotations...)
			}
			return nil
		})(path, info, err)
}

func (v *MetadataWalker) validateSingleYamlFile(path string, contents string) []*github.CheckRunAnnotation {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		var annotations []*github.CheckRunAnnotation
		if strings.Contains(path, "owner.info.yaml") {
			annotations = parseStrict(path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			annotations = parseStrict(path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			annotations = v.validateRepositoryFile(path, contents)
		} else {
			v.IgnoredWithReason[path] = "file is neither owner info, nor service nor repository"
			return nil
		}
		if lintAnnotation := v.checkFormatting(path, contents); lintAnnotation != nil {
			annotations = append(annotations, lintAnnotation)
		}
		return annotations
	} else {
		v.IgnoredWithReason[path] = "file is not a .yaml or not situated in owners/"
		return nil
	}
}

func parseStrict[T openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto](path string, contents string, resultPtr *T) []*github.CheckRunAnnotation {
	decoder := yaml.NewDecoder(strings.NewReader(contents))
	decoder.KnownFields(true)
	err := decoder.Decode(resultPtr)
	if err != nil {
		var yamlErrorMessages []string
		var typeErr *yaml.TypeError
		if errors.As(err, &typeErr) && typeErr != nil {
			yamlErrorMessages = typeErr.Errors
		} else {
			yamlErrorMessages = []string{err.Error()}
		}
		annotations, tErr := yamlErrorMessagesToAnnotations(path, yamlErrorMessages)
		if tErr != nil {
			return []*github.CheckRunAnnotation{
				errorToAnnotation(path, errors.Join(err, tErr)),
			}
		}
		return annotations
	}

	return nil
}

func yamlErrorMessagesToAnnotations(path string, errorMessages []string) ([]*github.CheckRunAnnotation, error) {
	result := make([]*github.CheckRunAnnotation, 0)
	yamlErrRegexp := regexp.MustCompile("(?:yaml: )?line ([0-9]+): (.*)")
	for _, ytem := range errorMessages {
		matches := yamlErrRegexp.FindStringSubmatch(ytem)
		if matches == nil || len(matches) != 3 {
			return nil, fmt.Errorf("failed to parse yaml type error messages from %s", ytem)
		}
		lineNum, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse yaml type error line number from %s: %w", ytem, err)
		}
		result = append(result, &github.CheckRunAnnotation{
			Path:            github.Ptr(path),
			StartLine:       github.Ptr(lineNum),
			EndLine:         github.Ptr(lineNum),
			AnnotationLevel: github.Ptr("failure"),
			Message:         github.Ptr(matches[2]),
		})

	}
	return result, nil
}

func errorToAnnotation(path string, err error) *github.CheckRunAnnotation {
	return &github.CheckRunAnnotation{
		Path:            github.Ptr(path),
		StartLine:       github.Ptr(1),
		EndLine:         github.Ptr(1),
		AnnotationLevel: github.Ptr("failure"),
		Message:         github.Ptr(err.Error()),
		Title:           github.Ptr("Unparsable Error in this file - unable to find exact line numbers"),
	}
}

func (v *MetadataWalker) validateRepositoryFile(path string, contents string) []*github.CheckRunAnnotation {
	repositoryDto := &openapi.RepositoryDto{}
	parseAnnotations := parseStrict(path, contents, repositoryDto)
	_, after, found := strings.Cut(path, "/repositories/")
	repoKey, isYaml := strings.CutSuffix(after, ".yaml")
	if found && isYaml {
		if annotation := v.checkKeyDuplication(path, repoKey); annotation != nil {
			parseAnnotations = append(parseAnnotations, annotation)
		}
		if annotation := v.checkUrlDuplication(path, contents); annotation != nil {
			parseAnnotations = append(parseAnnotations, annotation)
		}
	}

	return parseAnnotations
}

func (v *MetadataWalker) checkKeyDuplication(path string, repoKey string) *github.CheckRunAnnotation {
	var annotation *github.CheckRunAnnotation
	if otherFile, isDuplicatedKey := v.walkedRepos.keyToPath[repoKey]; isDuplicatedKey {
		annotation = &github.CheckRunAnnotation{
			Path:            github.Ptr(path),
			StartLine:       github.Ptr(1),
			EndLine:         github.Ptr(1),
			AnnotationLevel: github.Ptr("failure"),
			Message:         github.Ptr(fmt.Sprintf("Repository key already used by %s", otherFile)),
		}
	} else {
		v.walkedRepos.keyToPath[repoKey] = path
	}
	return annotation
}

func (v *MetadataWalker) checkUrlDuplication(path string, contents string) *github.CheckRunAnnotation {
	for lineNum, line := range strings.Split(contents, lineSeparatorCharacter) {
		if strings.HasPrefix(line, "url: ") {
			url := strings.TrimSpace(strings.ReplaceAll(line, "url: ", ""))
			if otherFile, isDuplicatedUrl := v.walkedRepos.urlToPath[url]; isDuplicatedUrl {
				return &github.CheckRunAnnotation{
					Path:            github.Ptr(path),
					StartLine:       github.Ptr(lineNum + 1),
					EndLine:         github.Ptr(lineNum + 1),
					AnnotationLevel: github.Ptr("failure"),
					Message:         github.Ptr(fmt.Sprintf("Repository url already used by %s", otherFile)),
				}
			} else {
				v.walkedRepos.urlToPath[url] = path
			}
		}
	}
	return nil
}

func (v *MetadataWalker) checkFormatting(path string, content string) *github.CheckRunAnnotation {
	formatted, err := v.fmtEngine.FormatContent([]byte(content))
	if err != nil {
		// ignoring marshalling errors because we already checked decoding with parseStrict
		return nil
	}
	diff := &yamlfmt.FormatDiff{
		Original:  content,
		Formatted: string(formatted),
		LineSep:   lineSeparatorCharacter,
	}
	if diff.Changed() {
		diffMsg, diffCount := diff.MultilineDiff()
		v.hasFormatErrors = true
		return &github.CheckRunAnnotation{
			Path:            github.Ptr(path),
			StartLine:       github.Ptr(1),
			EndLine:         github.Ptr(1),
			AnnotationLevel: github.Ptr("failure"),
			Message:         github.Ptr(diffMsg),
			Title:           github.Ptr(fmt.Sprintf("This file contains %d formatting errors.\nYou can use the \"Fix formatting\" action of this check to automatically reformat the files.", diffCount)),
		}
	}
	return nil
}
