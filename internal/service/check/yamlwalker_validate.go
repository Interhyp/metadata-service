package check

import (
	"errors"
	"fmt"
	"github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/go-git/go-billy/v5/util"
	"github.com/google/go-github/v70/github"
	"github.com/google/yamlfmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
)

const (
	annotationLevelWarning         = "warning"
	deploymentRepositoryIdentifier = "helm-deployment"
	refProtectionBranchIdentifier  = "branches"
	refProtectionTagIdentifier     = "tags"
)

func (v *MetadataWalker) ValidateMetadata() error {
	return util.Walk(v.fs, v.config.rootDir, v.validateWalkFunc)
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
		if annotation := v.checkMainlineProtection(path, repositoryDto); annotation != nil {
			parseAnnotations = append(parseAnnotations, annotation)
		}
		if annotations := v.checkRequiredConditions(path, repositoryDto); len(annotations) > 0 {
			parseAnnotations = append(parseAnnotations, annotations...)
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

func (v *MetadataWalker) checkMainlineProtection(path string, dto *openapi.RepositoryDto) *github.CheckRunAnnotation {
	if !v.config.requireMainlinePrProtection {
		return nil
	}
	if dto == nil ||
		dto.Configuration == nil ||
		dto.Configuration.RefProtections == nil ||
		dto.Configuration.RefProtections.Branches == nil {
		return nil
	}
	hasMainlineProtection := false
	for _, r := range dto.Configuration.RefProtections.Branches.RequirePR {
		hasMainlineProtection = hasMainlineProtection || r.Pattern == ":MAINLINE:"
	}
	if !hasMainlineProtection {
		return &github.CheckRunAnnotation{
			Path:            github.Ptr(path),
			StartLine:       github.Ptr(1),
			EndLine:         github.Ptr(1),
			AnnotationLevel: github.Ptr("warning"),
			Message:         github.Ptr("This file does not contain the requirePR mainline protection."),
			Title:           github.Ptr("mainline unprotected"),
		}
	}
	return nil
}

func (v *MetadataWalker) checkRequiredConditions(path string, dto *openapi.RepositoryDto) []*github.CheckRunAnnotation {
	if len(v.config.expectedExemptions) == 0 {
		return nil
	}
	if dto == nil || dto.Configuration == nil && (dto.Configuration.RequireConditions == nil && dto.Configuration.RefProtections == nil) {
		return nil
	}
	annotations := make([]*github.CheckRunAnnotation, 0)
	for _, expected := range v.config.expectedExemptions {
		if !strings.Contains(path, deploymentRepositoryIdentifier) {
			continue
		}
		conditionExists, missingConditionExemptions := v.checkExpectedExemptionOnRequiredConditions(expected, dto)
		refProtectionExists, missingRefProtectionExemptions := v.checkExpectedExemptionOnRefProtections(expected, dto)
		if !conditionExists && !refProtectionExists {
			annotations = append(annotations, &github.CheckRunAnnotation{
				Path:            github.Ptr(path),
				StartLine:       github.Ptr(1),
				EndLine:         github.Ptr(1),
				AnnotationLevel: github.Ptr(annotationLevelWarning),
				Message:         github.Ptr(fmt.Sprintf("This file does not contain the required condition/refProtection %s with the refMatcher %s.", expected.Name, expected.RefMatcher)),
				Title:           github.Ptr("missing expected condition/refProtection"),
			})
		}
		if len(missingConditionExemptions) > 0 || len(missingRefProtectionExemptions) > 0 {
			missing := append(missingConditionExemptions, missingRefProtectionExemptions...)
			v.hasMissingRequiredConditionExemptions = missing
			annotations = append(annotations, &github.CheckRunAnnotation{
				Path:            github.Ptr(path),
				StartLine:       github.Ptr(1),
				EndLine:         github.Ptr(1),
				AnnotationLevel: github.Ptr(annotationLevelWarning),
				Message:         github.Ptr(fmt.Sprintf("This file does not contain all required exemptions %s for condition %s with the refMatcher %s.", strings.Join(expected.Exemptions, ", "), expected.Name, expected.RefMatcher)),
				Title:           github.Ptr("missing expected required exemptions"),
			})
		}
	}

	return annotations
}

func (v *MetadataWalker) checkExpectedExemptionOnRequiredConditions(expected config.CheckedExpectedExemption, dto *openapi.RepositoryDto) (bool, []MissingRequiredConditionExemption) {
	missingExemptions := make([]MissingRequiredConditionExemption, 0)
	// requiredCondition missing false, exists true
	okResult := false
	if dto.Configuration.RequireConditions == nil && isExpectedExemptionCondition(expected) {
		return okResult, missingExemptions
	}
	if condition, ok := dto.Configuration.RequireConditions[expected.Name]; ok && condition.RefMatcher == expected.RefMatcher {
		okResult = ok
		if missing := allEntriesExist(expected.Exemptions, condition.Exemptions); len(missing) > 0 {
			missingExemptions = append(missingExemptions, MissingRequiredConditionExemption{
				Name:       expected.Name,
				RefMatcher: expected.RefMatcher,
				Exemptions: missing,
			})
		}
	}

	return okResult, missingExemptions
}

func (v *MetadataWalker) checkExpectedExemptionOnRefProtections(expected config.CheckedExpectedExemption, dto *openapi.RepositoryDto) (bool, []MissingRequiredConditionExemption) {
	missingExemptions := make([]MissingRequiredConditionExemption, 0)
	// requiredCondition missing false, exists true
	okResult := false
	if dto.Configuration.RefProtections == nil && !isExpectedExemptionCondition(expected) {
		return okResult, missingExemptions
	}
	var protectedRefs []openapi.ProtectedRef
	switch expected.Name {
	case fmt.Sprintf("%s.requirePR", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.RequirePR != nil {
			protectedRefs = dto.Configuration.RefProtections.Branches.RequirePR
		}
	case fmt.Sprintf("%s.preventAllChanges", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.PreventAllChanges != nil {

			protectedRefs = dto.Configuration.RefProtections.Branches.PreventAllChanges
		}
	case fmt.Sprintf("%s.preventCreation", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.PreventCreation != nil {
			protectedRefs = dto.Configuration.RefProtections.Branches.PreventCreation
		}
	case fmt.Sprintf("%s.preventDeletion", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.PreventDeletion != nil {
			protectedRefs = dto.Configuration.RefProtections.Branches.PreventDeletion
		}
	case fmt.Sprintf("%s.preventPush", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.PreventPush != nil {
			protectedRefs = dto.Configuration.RefProtections.Branches.PreventPush
		}
	case fmt.Sprintf("%s.preventForcePush", refProtectionBranchIdentifier):
		if dto.Configuration.RefProtections.Branches != nil && dto.Configuration.RefProtections.Branches.PreventForcePush != nil {
			protectedRefs = dto.Configuration.RefProtections.Branches.PreventForcePush
		}
	case fmt.Sprintf("%s.preventAllChanges", refProtectionTagIdentifier):
		if dto.Configuration.RefProtections.Tags != nil && dto.Configuration.RefProtections.Tags.PreventAllChanges != nil {
			protectedRefs = dto.Configuration.RefProtections.Tags.PreventAllChanges
		}
	case fmt.Sprintf("%s.preventCreation", refProtectionTagIdentifier):
		if dto.Configuration.RefProtections.Tags != nil && dto.Configuration.RefProtections.Tags.PreventCreation != nil {
			protectedRefs = dto.Configuration.RefProtections.Tags.PreventCreation
		}
	case fmt.Sprintf("%s.preventDeletion", refProtectionTagIdentifier):
		if dto.Configuration.RefProtections.Tags != nil && dto.Configuration.RefProtections.Tags.PreventDeletion != nil {
			protectedRefs = dto.Configuration.RefProtections.Tags.PreventDeletion
		}
	case fmt.Sprintf("%s.preventForcePush", refProtectionTagIdentifier):
		if dto.Configuration.RefProtections.Tags != nil && dto.Configuration.RefProtections.Tags.PreventForcePush != nil {
			protectedRefs = dto.Configuration.RefProtections.Tags.PreventForcePush
		}
	}

	for _, protectedRef := range protectedRefs {
		if protectedRef.Pattern == expected.RefMatcher {
			okResult = true
			if missing := allEntriesExist(expected.Exemptions, protectedRef.Exemptions); len(missing) > 0 {
				okResult = true
				missingExemptions = append(missingExemptions, MissingRequiredConditionExemption{
					Name:       expected.Name,
					RefMatcher: expected.RefMatcher,
					Exemptions: missing,
				})
			}
		}
	}

	return okResult, missingExemptions
}

func isExpectedExemptionCondition(expected config.CheckedExpectedExemption) bool {
	return !strings.Contains(expected.Name, "branches") || !strings.Contains(expected.Name, "tags")
}

func allEntriesExist(arr1, arr2 []string) []string {
	var missing []string
	for _, str1 := range arr1 {
		found := false
		for _, str2 := range arr2 {
			if str1 == str2 {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, str1)
		}
	}
	return missing
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
