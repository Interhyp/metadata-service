package check

import (
	"errors"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	gogithub "github.com/google/go-github/v69/github"
	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/engine"
	"github.com/google/yamlfmt/formatters/basic"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type walkedRepos struct {
	urlToPath map[string]string
	keyToPath map[string]string
}
type MetadataWalker struct {
	fs                billy.Filesystem
	Annotations       []*gogithub.CheckRunAnnotation
	Errors            map[string]error
	IgnoredWithReason map[string]string
	walkedRepos       walkedRepos
	fmtEngine         yamlfmt.Engine
	hasFormatErrors   bool
	rootDir           string
}

const lineBreakStyle = yamlfmt.LineBreakStyleLF
const lineSeparatorCharacter = "\n"

func MetadataYamlFileWalker(filesys billy.Filesystem) *MetadataWalker {
	c := &basic.Config{ //https://github.com/google/yamlfmt/blob/main/docs/config-file.md
		Indent:                 2,
		LineEnding:             lineBreakStyle,
		PadLineComments:        1,
		RetainLineBreaksSingle: true,
		ScanFoldedAsLiteral:    true,
	}
	b := &basic.BasicFormatter{
		Config:       c,
		Features:     basic.ConfigureFeaturesFromConfig(c),
		YAMLFeatures: basic.ConfigureYAMLFeaturesFromConfig(c),
	}
	eng := &engine.ConsecutiveEngine{
		LineSepCharacter: lineSeparatorCharacter,
		Formatter:        b,
		Quiet:            false,
		ContinueOnError:  true,
		OutputFormat:     engine.EngineOutputDefault,
	}
	validator := MetadataWalker{
		fs:                filesys,
		Annotations:       make([]*gogithub.CheckRunAnnotation, 0),
		Errors:            make(map[string]error),
		IgnoredWithReason: make(map[string]string),
		walkedRepos: walkedRepos{
			urlToPath: make(map[string]string),
			keyToPath: make(map[string]string),
		},
		fmtEngine: eng,
		rootDir:   "/",
	}
	return &validator
}
func (v *MetadataWalker) WithRootDir(newRoot string) *MetadataWalker {
	v.rootDir = newRoot
	return v
}

func (v *MetadataWalker) ValidateMetadata() error {
	return util.Walk(v.fs, v.rootDir, v.validateWalkFunc)
}

func (v *MetadataWalker) validateWalkFunc(path string, info fs.FileInfo, err error) error {
	return v.walkFunc(
		func(fileContents []byte) error {
			trimmed := strings.Trim(path, "/")
			annotations := v.validateYamlFile(trimmed, string(fileContents))
			if len(annotations) > 0 {
				v.Annotations = append(v.Annotations, annotations...)
			}
			return nil
		})(path, info, err)
}

func (v *MetadataWalker) FormatMetadata() error {
	return util.Walk(v.fs, v.rootDir, v.formatWalkFunc)
}

func (v *MetadataWalker) formatWalkFunc(path string, info fs.FileInfo, err error) error {
	return v.walkFunc(
		func(fileContents []byte) error {
			return v.formatYamlFile(fileContents, path)
		})(path, info, err)
}

func (v *MetadataWalker) walkFunc(fileFunc func(fileContents []byte) error) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		// we do not want to return errors to walk through all available files
		if err != nil {
			v.Errors[path] = err
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		fileContents, err := util.ReadFile(v.fs, path)
		if err != nil {
			v.Errors[path] = err
			return nil
		}

		err = fileFunc(fileContents)
		if err != nil {
			v.Errors[path] = err
		}

		return nil
	}
}

func (v *MetadataWalker) formatYamlFile(fileContents []byte, path string) error {
	formatted, err := v.fmtEngine.FormatContent(fileContents)
	if err != nil {
		return err
	}
	if len(formatted) == 0 {
		return fmt.Errorf("missing formatter result")
	}
	return v.replaceFileContent(path, formatted)
}

func (v *MetadataWalker) replaceFileContent(path string, formatted []byte) error {
	err := v.fs.Remove(path)
	if err != nil {
		return err
	}

	f, err := v.fs.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}

	_, err = f.Write(formatted)
	if err != nil {
		return err
	}
	return nil
}

func (v *MetadataWalker) validateYamlFile(path string, contents string) []*gogithub.CheckRunAnnotation {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		var annotations []*gogithub.CheckRunAnnotation
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
		if lintAnnotation := v.lint(path, contents); lintAnnotation != nil {
			annotations = append(annotations, lintAnnotation)
		}
		return annotations
	} else {
		v.IgnoredWithReason[path] = "file is not a .yaml or not situated in owners/"
		return nil
	}
}

func parseStrict[T openapi.OwnerDto | openapi.ServiceDto | openapi.RepositoryDto](path string, contents string, resultPtr *T) []*gogithub.CheckRunAnnotation {
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
			return []*gogithub.CheckRunAnnotation{
				errorToAnnotation(path, errors.Join(err, tErr)),
			}
		}
		return annotations
	}

	return nil
}

func errorToAnnotation(path string, err error) *gogithub.CheckRunAnnotation {
	return &gogithub.CheckRunAnnotation{
		Path:            gogithub.Ptr(path),
		StartLine:       gogithub.Ptr(1),
		EndLine:         gogithub.Ptr(1),
		AnnotationLevel: gogithub.Ptr("failure"),
		Message:         gogithub.Ptr(err.Error()),
		Title:           gogithub.Ptr("Unparsable Error in this file - unable to find exact line numbers"),
	}
}

func yamlErrorMessagesToAnnotations(path string, errorMessages []string) ([]*gogithub.CheckRunAnnotation, error) {
	result := make([]*gogithub.CheckRunAnnotation, 0)
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
		result = append(result, &gogithub.CheckRunAnnotation{
			Path:            gogithub.Ptr(path),
			StartLine:       gogithub.Ptr(lineNum),
			EndLine:         gogithub.Ptr(lineNum),
			AnnotationLevel: gogithub.Ptr("failure"),
			Message:         gogithub.Ptr(matches[2]),
		})

	}
	return result, nil
}

func (v *MetadataWalker) validateRepositoryFile(path string, contents string) []*gogithub.CheckRunAnnotation {
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

func (v *MetadataWalker) lint(path string, content string) *gogithub.CheckRunAnnotation {
	formatted, err := v.fmtEngine.FormatContent([]byte(content))
	if err != nil {
		return &gogithub.CheckRunAnnotation{
			Path:            gogithub.Ptr(path),
			StartLine:       gogithub.Ptr(1),
			EndLine:         gogithub.Ptr(1),
			AnnotationLevel: gogithub.Ptr("failure"),
			Message:         gogithub.Ptr(err.Error()),
			Title:           gogithub.Ptr("Formatting error"),
		}
	}
	diff := &yamlfmt.FormatDiff{
		Original:  content,
		Formatted: string(formatted),
		LineSep:   lineSeparatorCharacter,
	}
	if diff.Changed() {
		diffMsg, diffCount := diff.MultilineDiff()
		v.hasFormatErrors = true
		return &gogithub.CheckRunAnnotation{
			Path:            gogithub.Ptr(path),
			StartLine:       gogithub.Ptr(1),
			EndLine:         gogithub.Ptr(1),
			AnnotationLevel: gogithub.Ptr("failure"),
			Message:         gogithub.Ptr(diffMsg),
			Title:           gogithub.Ptr(fmt.Sprintf("This file contains %d formatting errors.\nYou can use the \"Fix formatting\" action of this check to automatically reformat the files.", diffCount)),
		}
	}
	return nil
}

func (v *MetadataWalker) checkKeyDuplication(path string, repoKey string) *gogithub.CheckRunAnnotation {
	var annotation *gogithub.CheckRunAnnotation
	if otherFile, isDuplicatedKey := v.walkedRepos.keyToPath[repoKey]; isDuplicatedKey {
		annotation = &gogithub.CheckRunAnnotation{
			Path:            gogithub.Ptr(path),
			StartLine:       gogithub.Ptr(1),
			EndLine:         gogithub.Ptr(1),
			AnnotationLevel: gogithub.Ptr("failure"),
			Message:         gogithub.Ptr(fmt.Sprintf("Repository key already used by %s", otherFile)),
		}
	} else {
		v.walkedRepos.keyToPath[repoKey] = path
	}
	return annotation
}

func (v *MetadataWalker) checkUrlDuplication(path string, contents string) *gogithub.CheckRunAnnotation {
	for lineNum, line := range strings.Split(contents, lineSeparatorCharacter) {
		if strings.HasPrefix(line, "url: ") {
			url := strings.TrimSpace(strings.ReplaceAll(line, "url: ", ""))
			if otherFile, isDuplicatedUrl := v.walkedRepos.urlToPath[url]; isDuplicatedUrl {
				return &gogithub.CheckRunAnnotation{
					Path:            gogithub.Ptr(path),
					StartLine:       gogithub.Ptr(lineNum + 1),
					EndLine:         gogithub.Ptr(lineNum + 1),
					AnnotationLevel: gogithub.Ptr("failure"),
					Message:         gogithub.Ptr(fmt.Sprintf("Repository url already used by %s", otherFile)),
				}
			} else {
				v.walkedRepos.urlToPath[url] = path
			}
		}
	}
	return nil
}
