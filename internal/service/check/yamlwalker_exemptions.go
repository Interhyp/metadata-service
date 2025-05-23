package check

import (
	openapi "github.com/Interhyp/metadata-service/api"
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/go-git/go-billy/v5/util"
	"gopkg.in/yaml.v3"
	"io/fs"
	"strings"
)

func (v *MetadataWalker) FixExemptions() error {
	return util.Walk(v.fs, v.config.rootDir, v.fixExemptionsFunc)
}

func (v *MetadataWalker) fixExemptionsFunc(path string, info fs.FileInfo, err error) error {
	return v.walkFunc(
		func(fileContents []byte) error {
			return v.fixExemptionsInFile(fileContents, path)
		})(path, info, err)
}

func (v *MetadataWalker) fixExemptionsInFile(fileContents []byte, path string) error {
	if strings.Contains(path, "/repositories/") {
		v.validateRepositoryFile(path, string(fileContents))
		if len(v.hasMissingRequiredConditionExemptions) > 0 {
			dto := &openapi.RepositoryDto{}
			_ = parseStrict(path, string(fileContents), dto)
			for _, missingExemption := range v.hasMissingRequiredConditionExemptions {
				switch missingExemption.Name {
				case "branches.requirePR":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.RequirePR {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.RequirePR = newRefProtection
				case "branches.preventAllChanges":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.PreventAllChanges {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.PreventAllChanges = newRefProtection
				case "branches.preventCreation":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.PreventCreation {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.PreventCreation = newRefProtection
				case "branches.preventDeletion":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.PreventDeletion {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.PreventDeletion = newRefProtection
				case "branches.preventPush":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.PreventPush {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.PreventPush = newRefProtection
				case "branches.preventForcePush":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Branches.PreventForcePush {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Branches.PreventForcePush = newRefProtection
				case "tags.preventAllChanges":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Tags.PreventAllChanges {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Tags.PreventAllChanges = newRefProtection
				case "tags.preventCreation":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Tags.PreventCreation {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Tags.PreventCreation = newRefProtection
				case "tags.preventDeletion":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Tags.PreventDeletion {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Tags.PreventDeletion = newRefProtection
				case "tags.preventForcePush":
					var newRefProtection []openapi.ProtectedRef
					for _, refProtection := range dto.Configuration.RefProtections.Tags.PreventForcePush {
						newRefProtection = addMissingExemption(refProtection, missingExemption, newRefProtection)
					}
					dto.Configuration.RefProtections.Tags.PreventForcePush = newRefProtection
				}
				if isExpectedRequiredCondition(config.CheckedRequiredConditions{Name: missingExemption.Name}) {
					if current, ok := dto.Configuration.RequireConditions[missingExemption.Name]; ok && current.RefMatcher == missingExemption.RefMatcher {
						current.Exemptions = append(current.Exemptions, missingExemption.Exemptions...)
						dto.Configuration.RequireConditions[missingExemption.Name] = current
					}
				}
			}
			fixed, err := yaml.Marshal(dto)
			if err != nil {
				return err
			}
			return v.formatSingleYamlFile(fixed, path)
		}
	}
	return nil
}

func addMissingExemption(refProtection openapi.ProtectedRef, missingExemption MissingRequiredConditionExemption, newRefProtection []openapi.ProtectedRef) []openapi.ProtectedRef {
	if refProtection.Pattern == missingExemption.RefMatcher {
		refProtection.Exemptions = append(refProtection.Exemptions, missingExemption.Exemptions...)
	}
	return append(newRefProtection, refProtection)
}
