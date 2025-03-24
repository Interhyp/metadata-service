package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMetadataRepoName_MetadataRepoUrl(t *testing.T) {
	_, customCfg := New()
	cut := customCfg.(*CustomConfigImpl)
	cut.VMetadataRepoUrl = "https://some-domain.com/bitbucket/scm/tmpl/service-metadata.git"
	repoName := cut.MetadataRepoName()
	require.Equal(t, "service-metadata", repoName)

	cut.VMetadataRepoUrl = "https://github.com/some-org/service-metadata-test.git"
	repoName = cut.MetadataRepoName()
	require.Equal(t, "service-metadata-test", repoName)
}

func TestMetadataRepoProject_MetadataRepoUrl(t *testing.T) {
	_, customCfg := New()
	cut := customCfg.(*CustomConfigImpl)
	cut.VMetadataRepoUrl = "https://some-domain.com/bitbucket/scm/tmpl/service-metadata.git"
	repoName := cut.MetadataRepoProject()
	require.Equal(t, "tmpl", repoName)

	cut.VMetadataRepoUrl = "https://github.com/some-org/service-metadata-test.git"
	repoName = cut.MetadataRepoProject()
	require.Equal(t, "some-org", repoName)
}
