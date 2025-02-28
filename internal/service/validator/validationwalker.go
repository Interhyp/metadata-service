package validator

import (
	"errors"
	"fmt"
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	gogithub "github.com/google/go-github/v69/github"
	"gopkg.in/yaml.v3"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
)

type ValidationWalker struct {
	fs                billy.Filesystem
	Annotations       []*gogithub.CheckRunAnnotation
	Errors            map[string]error
	IgnoredWithReason map[string]string
	repoUrlsToKey     map[string]string
}

func NewValidationWalker(filesys billy.Filesystem, repos openapi.RepositoryListDto) *ValidationWalker {
	validator := ValidationWalker{
		fs:                filesys,
		Annotations:       make([]*gogithub.CheckRunAnnotation, 0),
		Errors:            make(map[string]error),
		repoUrlsToKey:     make(map[string]string),
		IgnoredWithReason: make(map[string]string),
	}
	for repoKey, repo := range repos.Repositories {
		validator.repoUrlsToKey[repo.Url] = repoKey
	}
	return &validator
}

func (v *ValidationWalker) WalkerFunc(path string, info fs.FileInfo, err error) error {
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

	f, err := util.ReadFile(v.fs, path)
	if err != nil {
		v.Errors[path] = err
	}

	trimmed := strings.Trim(path, "/")
	annotations := v.validateYamlFile(trimmed, string(f))
	if len(annotations) > 0 {
		v.Annotations = append(v.Annotations, annotations...)
	}

	return nil
}

func (v *ValidationWalker) validateYamlFile(path string, contents string) []*gogithub.CheckRunAnnotation {
	if strings.HasPrefix(path, "owners/") && strings.HasSuffix(path, ".yaml") {
		if strings.Contains(path, "owner.info.yaml") {
			return parseStrict(path, contents, &openapi.OwnerDto{})
		} else if strings.Contains(path, "/services/") {
			return parseStrict(path, contents, &openapi.ServiceDto{})
		} else if strings.Contains(path, "/repositories/") {
			return v.validateRepositoryFile(path, contents)
		} else {
			v.IgnoredWithReason[path] = "file is neither owner info, nor service nor repository"
			return nil
		}
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

func (v *ValidationWalker) validateRepositoryFile(path string, contents string) []*gogithub.CheckRunAnnotation {
	repositoryDto := &openapi.RepositoryDto{}
	parseAnnotations := parseStrict(path, contents, repositoryDto)
	_, after, found := strings.Cut(path, "/repositories/")
	repoKey, isYaml := strings.CutSuffix(after, ".yaml")
	if found && isYaml {
		for lineNum, line := range strings.Split(contents, "\n") {
			if strings.HasPrefix(line, "url: ") {
				url := strings.TrimSpace(strings.ReplaceAll(line, "url: ", ""))
				if otherRepoKey, found := v.repoUrlsToKey[url]; found && otherRepoKey != repoKey {
					parseAnnotations = append(parseAnnotations, &gogithub.CheckRunAnnotation{
						Path: gogithub.Ptr(path),
						// line numbers start with 1
						// see https://docs.github.com/en/enterprise-cloud@latest/rest/checks/runs?apiVersion=2022-11-28#update-a-check-run
						StartLine:       gogithub.Ptr(lineNum + 1),
						EndLine:         gogithub.Ptr(lineNum + 1),
						AnnotationLevel: gogithub.Ptr("failure"),
						Message:         gogithub.Ptr(fmt.Sprintf("Repository url already used by %s", otherRepoKey)),
					})
				}
			}
		}
	}

	return parseAnnotations
}
