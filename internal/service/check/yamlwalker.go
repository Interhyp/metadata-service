package check

import (
	"github.com/Interhyp/metadata-service/internal/acorn/config"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"github.com/google/go-github/v70/github"
	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/engine"
	"github.com/google/yamlfmt/formatters/basic"
	"io/fs"
	"path/filepath"
	"strings"
)

type walkedRepos struct {
	urlToPath map[string]string
	keyToPath map[string]string
}
type MetadataWalker struct {
	fs                billy.Filesystem
	Annotations       []*github.CheckRunAnnotation
	Errors            map[string]error
	IgnoredWithReason map[string]string
	walkedRepos       walkedRepos
	fmtEngine         yamlfmt.Engine
	hasFormatErrors   bool
	config            Config
}

type Config struct {
	rootDir                     string
	indentation                 int
	requireMainlinePrProtection bool
	expectedRequiredConditions  []config.CheckedRequiredConditions
}

type Option = func(config *Config)

func WithRootDir(rootDir string) Option {
	return func(config *Config) {
		config.rootDir = rootDir
	}
}

func WithIndentation(indentation int) Option {
	return func(config *Config) {
		config.indentation = indentation
	}
}

func WithMainlinePrProtection(requireMainlinePrProtection bool) Option {
	return func(config *Config) {
		config.requireMainlinePrProtection = requireMainlinePrProtection
	}
}

func WithExpectedRequiredConditions(expectedReqConditions []config.CheckedRequiredConditions) Option {
	return func(config *Config) {
		config.expectedRequiredConditions = expectedReqConditions
	}
}

const lineBreakStyle = yamlfmt.LineBreakStyleLF
const lineSeparatorCharacter = "\n"

func MetadataYamlFileWalker(filesys billy.Filesystem, options ...Option) *MetadataWalker {
	walkerConf := Config{
		rootDir:                     "/",
		indentation:                 2,
		requireMainlinePrProtection: false,
		expectedRequiredConditions:  nil,
	}
	for _, option := range options {
		option(&walkerConf)
	}
	fmtConf := &basic.Config{ //https://github.com/google/yamlfmt/blob/main/docs/config-file.md
		Indent:                 walkerConf.indentation,
		LineEnding:             lineBreakStyle,
		PadLineComments:        1,
		RetainLineBreaksSingle: true,
		ScanFoldedAsLiteral:    true,
	}
	formatter := &basic.BasicFormatter{
		Config:       fmtConf,
		Features:     basic.ConfigureFeaturesFromConfig(fmtConf),
		YAMLFeatures: basic.ConfigureYAMLFeaturesFromConfig(fmtConf),
	}
	fmtEngine := &engine.ConsecutiveEngine{
		LineSepCharacter: lineSeparatorCharacter,
		Formatter:        formatter,
		Quiet:            false,
		ContinueOnError:  true,
		OutputFormat:     engine.EngineOutputDefault,
	}
	validator := MetadataWalker{
		fs:                filesys,
		Annotations:       make([]*github.CheckRunAnnotation, 0),
		Errors:            make(map[string]error),
		IgnoredWithReason: make(map[string]string),
		walkedRepos: walkedRepos{
			urlToPath: make(map[string]string),
			keyToPath: make(map[string]string),
		},
		fmtEngine: fmtEngine,
		config:    walkerConf,
	}
	return &validator
}

func (v *MetadataWalker) walkFunc(perFileFunc func(fileContents []byte) error) filepath.WalkFunc {
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

		err = perFileFunc(fileContents)
		if err != nil {
			v.Errors[path] = err
		}

		return nil
	}
}
