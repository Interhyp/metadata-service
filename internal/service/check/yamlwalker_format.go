package check

import (
	"fmt"
	"github.com/go-git/go-billy/v5/util"
	"io/fs"
)

func (v *MetadataWalker) FormatMetadata() error {
	return util.Walk(v.fs, v.config.rootDir, v.formatWalkFunc)
}

func (v *MetadataWalker) formatWalkFunc(path string, info fs.FileInfo, err error) error {
	return v.walkFunc(
		func(fileContents []byte) error {
			return v.formatSingleYamlFile(fileContents, path)
		})(path, info, err)
}

func (v *MetadataWalker) formatSingleYamlFile(fileContents []byte, path string) error {
	formatted, err := v.fmtEngine.FormatContent(fileContents)
	if err != nil {
		return err
	}
	if len(formatted) == 0 {
		return fmt.Errorf("missing formatter result")
	}
	return v.replaceFileContent(path, formatted)
}

func (v *MetadataWalker) replaceFileContent(path string, newContent []byte) error {
	err := v.fs.Remove(path)
	if err != nil {
		return err
	}

	f, err := v.fs.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}

	_, err = f.Write(newContent)
	if err != nil {
		return err
	}
	return nil
}
