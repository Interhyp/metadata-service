package checkoutmock

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

const ownerInfo = `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
defaultJiraProject: ISSUE
groups:
  users:
    - some-other-user
    - a-very-special-user
`

const ownerInfoNoPromoters = `contact: somebody@some-organisation.com
teamsChannelURL: https://teams.microsoft.com/l/channel/somechannel
productOwner: kschlangenheldt
defaultJiraProject: ISSUE
`

const service = `quicklinks:
- title: Swagger UI
  url: /swagger-ui/index.html
repositories:
- some-service-backend/helm-deployment
- some-service-backend/implementation
alertTarget: https://webhook.com/9asdflk29d4m39g
`

const deployment = `mainline: main
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/some-service-backend-deployment.git
deployment:
  kubernetes:
    instances:
    - namespace: project
      environment: prod
      cluster: openshift
    - namespace: project
      environment: dev
      cluster: openshift
    - namespace: project
      environment: test
      cluster: openshift
    - namespace: project
      environment: livetest
      cluster: openshift
generator: third-party-software
configuration:
  accessKeys:
  - key: DEPLOYMENT
    permission: REPO_READ
  - data: 'ssh-key abcdefgh.....'
    permission: REPO_WRITE
  commitMessageType: DEFAULT
  mergeConfig:
    defaultStrategy:
      id: "no-ff"
    strategies:
      - id: "no-ff"
      - id: "ff"
      - id: "ff-only"
      - id: "squash"
  requireIssue: true
  approvers:
    testing:
    - some-user
`
const expandableGroupsService = `quicklinks:
- title: Swagger UI
  url: /swagger-ui/index.html
repositories:
- some-service-backend-with-expandable-groups/helm-deployment
- some-service-backend/implementation
alertTarget: https://webhook.com/9asdflk29d4m39g
`

const expandableGroupsDeployment = `mainline: main
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/some-service-backend-with-expandable-groups-deployment.git
deployment:
  kubernetes:
    instances:
    - namespace: project
      environment: prod
      cluster: openshift
    - namespace: project
      environment: dev
      cluster: openshift
    - namespace: project
      environment: test
      cluster: openshift
    - namespace: project
      environment: livetest
      cluster: openshift
generator: third-party-software
configuration:
  accessKeys:
  - key: DEPLOYMENT
    permission: REPO_READ
  - data: 'ssh-key abcdefgh.....'
    permission: REPO_WRITE
  commitMessageType: DEFAULT
  mergeConfig:
    defaultStrategy:
      id: "no-ff"
    strategies:
      - id: "no-ff"
      - id: "ff"
      - id: "ff-only"
      - id: "squash"
  requireIssue: true
  watchers:
    - '@some-owner.users'
  refProtections:
    branches:
      requirePR:
        - pattern: ':MAINLINE:'
          exemptions:
            - '@some-owner.users'
  approvers:
    testing:
    - '@some-owner.users'
`

const deployment2 = `mainline: main
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/whatever-deployment.git
generator: third-party-software
`

const implementation = `mainline: master
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/some-service-backend.git
generator: java-spring-cloud
`

const implementation2 = `mainline: master
url: ssh://git@bitbucket.some-organisation.com:7999/PROJECT/whatever.git
generator: java-spring-cloud
`

const chart = `url: ssh://git@bitbucket.some-organisation.com:7999/helm/karma-wrapper.git
mainline: master
configuration:
  branchNameRegex: testing_.*
`

func writeFile(fs billy.Filesystem, filename string, contents string) error {
	f, err := fs.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(contents))
	return err
}

func New() (billy.Filesystem, error) {
	fs := memfs.New()
	err := fs.MkdirAll("owners/some-owner/services", 0755)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll("owners/some-owner/repositories", 0755)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll("owners/deleteme/services", 0755)
	if err != nil {
		return nil, err
	}
	err = fs.MkdirAll("owners/deleteme/repositories", 0755)
	if err != nil {
		return nil, err
	}

	err = writeFile(fs, "owners/some-owner/owner.info.yaml", ownerInfo)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/services/some-service-backend.yaml", service)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/services/some-service-backend-with-expandable-groups.yaml", expandableGroupsService)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/some-service-backend.helm-deployment.yaml", deployment)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/some-service-backend-with-expandable-groups.helm-deployment.yaml", expandableGroupsDeployment)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/some-service-backend.implementation.yaml", implementation)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/whatever.helm-deployment.yaml", deployment2)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/whatever.implementation.yaml", implementation2)
	if err != nil {
		return nil, err
	}
	err = writeFile(fs, "owners/some-owner/repositories/karma-wrapper.helm-chart.yaml", chart)
	if err != nil {
		return nil, err
	}

	err = writeFile(fs, "owners/deleteme/owner.info.yaml", ownerInfoNoPromoters)
	if err != nil {
		return nil, err
	}

	return fs, nil
}
