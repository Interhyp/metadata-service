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

func TestMetadataRepoName_SshMetadataRepoUrl(t *testing.T) {
	_, customCfg := New()
	cut := customCfg.(*CustomConfigImpl)
	cut.VSSHMetadataRepoUrl = "ssh://git@some-domain.com:7999/tmpl/service-metadata.git"
	repoName := cut.MetadataRepoName()
	require.Equal(t, "service-metadata", repoName)

	cut.VSSHMetadataRepoUrl = "git@github.com:some-org/service-metadata-test.git"
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

func TestMetadataRepoProject_SshMetadataRepoUrl(t *testing.T) {
	_, customCfg := New()
	cut := customCfg.(*CustomConfigImpl)
	cut.VSSHMetadataRepoUrl = "ssh://git@some-domain.com:7999/tmpl/service-metadata.git"
	repoName := cut.MetadataRepoProject()
	require.Equal(t, "tmpl", repoName)

	cut.VSSHMetadataRepoUrl = "git@github.com:some-org/service-metadata-test.git"
	repoName = cut.MetadataRepoProject()
	require.Equal(t, "some-org", repoName)
}
