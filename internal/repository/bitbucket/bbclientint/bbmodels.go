package bbclientint

// not part of spec - FileOrDirectory is missing useful fields

type PaginatedLines struct {
	Lines []struct {
		Text string `json:"text"`
	} `json:"lines"`
	Start         int  `json:"start"`
	Size          int  `json:"size"`
	IsLastPage    bool `json:"isLastPage"`
	Limit         int  `json:"limit"`
	NextPageStart *int `json:"nextPageStart"`
}

// part of spec, sorted alphabetically

type Changes struct {
	FromHash      string   `json:"fromHash"`
	ToHash        string   `json:"toHash"`
	Values        []Change `json:"values,omitempty"`
	Size          int      `json:"size"`
	IsLastPage    bool     `json:"isLastPage"`
	Start         int      `json:"start"`
	Limit         int      `json:"limit"`
	NextPageStart *int     `json:"nextPageStart"`
}

type Change struct {
	ContentId     string `json:"contentId"`
	FromContentId string `json:"fromContentId"`
	Path          struct {
		Components []string `json:"components"`
		Parent     string   `json:"parent"`
		Name       string   `json:"name"`
		Extension  string   `json:"extension"`
		ToString   string   `json:"toString"`
	} `json:"path"`
	Executable       bool   `json:"executable"`
	PercentUnchanged int    `json:"percentUnchanged"`
	Type             string `json:"type"`
	NodeType         string `json:"nodeType"`
	SrcExecutable    bool   `json:"srcExecutable"`
	Links            struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
	Properties struct {
		GitChangeType string `json:"gitChangeType"`
	} `json:"properties"`
}

type CommitBuildStatusRequest struct {
	Key         string       `yaml:"key" json:"key"`
	State       string       `yaml:"state" json:"state"`
	Url         string       `yaml:"url" json:"url"`
	BuildNumber *int32       `yaml:"buildNumber,omitempty" json:"buildNumber,omitempty"`
	Description *string      `yaml:"description,omitempty" json:"description,omitempty"`
	Duration    *int32       `yaml:"duration,omitempty" json:"duration,omitempty"`
	LastUpdated *int32       `yaml:"lastUpdated,omitempty" json:"lastUpdated,omitempty"`
	Name        *string      `yaml:"name,omitempty" json:"name,omitempty"`
	Parent      *string      `yaml:"parent,omitempty" json:"parent,omitempty"`
	Ref         *string      `yaml:"ref,omitempty" json:"ref,omitempty"`
	TestResults *TestResults `yaml:"testResults,omitempty" json:"testResults,omitempty"`
}

type Link struct {
	Href string  `yaml:"href" json:"href"`
	Name *string `yaml:"name,omitempty" json:"name,omitempty"`
}

type ProjectLinks struct {
	Self []Link `yaml:"self,omitempty" json:"self,omitempty"`
}

type PullRequest struct {
	Id           int64            `yaml:"id" json:"id"`
	Version      *int32           `yaml:"version,omitempty" json:"version,omitempty"`
	Title        string           `yaml:"title" json:"title"`
	Description  string           `yaml:"description" json:"description"`
	State        PullRequestState `yaml:"state" json:"state"`
	Open         bool             `yaml:"open" json:"open"`
	Closed       bool             `yaml:"closed" json:"closed"`
	CreatedDate  *int64           `yaml:"createdDate,omitempty" json:"createdDate,omitempty"`
	UpdatedDate  *int64           `yaml:"updatedDate,omitempty" json:"updatedDate,omitempty"`
	FromRef      RepositoryRef    `yaml:"fromRef" json:"fromRef"`
	ToRef        RepositoryRef    `yaml:"toRef" json:"toRef"`
	Locked       bool             `yaml:"locked" json:"locked"`
	Author       *UserRole        `yaml:"author,omitempty" json:"author,omitempty"`
	Reviewers    []UserRole       `yaml:"reviewers,omitempty" json:"reviewers,omitempty"`
	Participants []UserRole       `yaml:"participants,omitempty" json:"participants,omitempty"`
	Links        *ProjectLinks    `yaml:"links,omitempty" json:"links,omitempty"`
}

type PullRequestComment struct {
	Properties          PullRequestCommentProperties          `yaml:"properties" json:"properties"`
	Id                  int32                                 `yaml:"id" json:"id"`
	Version             int32                                 `yaml:"version" json:"version"`
	Text                string                                `yaml:"text" json:"text"`
	Author              PullRequestCommentAuthor              `yaml:"author" json:"author"`
	CreatedDate         int64                                 `yaml:"createdDate" json:"createdDate"`
	UpdatedDate         int64                                 `yaml:"updatedDate" json:"updatedDate"`
	Comments            []PullRequestComment                  `yaml:"comments" json:"comments"`
	Tasks               []interface{}                         `yaml:"tasks" json:"tasks"`
	Severity            string                                `yaml:"severity" json:"severity"`
	State               string                                `yaml:"state" json:"state"`
	PermittedOperations PullRequestCommentPermittedOperations `yaml:"permittedOperations" json:"permittedOperations"`
}

type PullRequestCommentAuthor struct {
	Name         string `yaml:"name" json:"name"`
	EmailAddress string `yaml:"emailAddress" json:"emailAddress"`
	Id           int32  `yaml:"id" json:"id"`
	DisplayName  string `yaml:"displayName" json:"displayName"`
	Active       bool   `yaml:"active" json:"active"`
	Slug         string `yaml:"slug" json:"slug"`
	Type         string `yaml:"type" json:"type"`
}

type PullRequestCommentPage struct {
	Size          int32                `yaml:"size" json:"size"`
	Limit         int32                `yaml:"limit" json:"limit"`
	Start         int32                `yaml:"start" json:"start"`
	IsLastPage    bool                 `yaml:"isLastPage" json:"isLastPage"`
	NextPageStart *int32               `yaml:"nextPageStart,omitempty" json:"nextPageStart,omitempty"`
	Values        []PullRequestComment `yaml:"values" json:"values"`
}

type PullRequestCommentPermittedOperations struct {
	Editable  bool `yaml:"editable" json:"editable"`
	Deletable bool `yaml:"deletable" json:"deletable"`
}

type PullRequestCommentProperties struct {
	Key string `yaml:"key" json:"key"`
}

type PullRequestCommentRequest struct {
	Text     string                           `yaml:"text" json:"text"`
	Parent   *PullRequestCommentRequestParent `yaml:"parent,omitempty" json:"parent,omitempty"`
	Anchor   *PullRequestCommentRequestAnchor `yaml:"anchor,omitempty" json:"anchor,omitempty"`
	Severity *string                          `yaml:"severity,omitempty" json:"severity,omitempty"`
	State    *string                          `yaml:"state,omitempty" json:"state,omitempty"`
}

type PullRequestCommentRequestAnchor struct {
	Line     *int32  `yaml:"line,omitempty" json:"line,omitempty"`
	LineType *string `yaml:"lineType,omitempty" json:"lineType,omitempty"`
	FileType *string `yaml:"fileType,omitempty" json:"fileType,omitempty"`
	Path     *string `yaml:"path,omitempty" json:"path,omitempty"`
	SrcPath  *string `yaml:"srcPath,omitempty" json:"srcPath,omitempty"`
}

type PullRequestCommentRequestParent struct {
	Id       int32   `yaml:"id" json:"id"`
	Severity *string `yaml:"severity,omitempty" json:"severity,omitempty"`
	State    *string `yaml:"state,omitempty" json:"state,omitempty"`
}

type PullRequestState string

// List of pullRequestState
const (
	OPEN     PullRequestState = "OPEN"
	MERGED   PullRequestState = "MERGED"
	DECLINED PullRequestState = "DECLINED"
)

// All allowed values of PullRequestState enum
var AllowedPullRequestStateEnumValues = []PullRequestState{
	"OPEN",
	"MERGED",
	"DECLINED",
}

type RepositoryRef struct {
	Id           string                  `yaml:"id" json:"id"`
	LatestCommit string                  `yaml:"latestCommit" json:"latestCommit"`
	Repository   RepositoryRefRepository `yaml:"repository" json:"repository"`
}

type RepositoryRefRepository struct {
	Slug    string                         `yaml:"slug" json:"slug"`
	Name    *string                        `yaml:"name,omitempty" json:"name,omitempty"`
	Project RepositoryRefRepositoryProject `yaml:"project" json:"project"`
}

type RepositoryRefRepositoryProject struct {
	Key string `yaml:"key" json:"key"`
}

type TestResults struct {
	Failed     int32 `yaml:"failed" json:"failed"`
	Skipped    int32 `yaml:"skipped" json:"skipped"`
	Successful int32 `yaml:"successful" json:"successful"`
}

type User struct {
	Id           *int32  `yaml:"id,omitempty" json:"id,omitempty"`
	Name         string  `yaml:"name" json:"name"`
	EmailAddress *string `yaml:"emailAddress,omitempty" json:"emailAddress,omitempty"`
	DisplayName  *string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Active       bool    `yaml:"active" json:"active"`
	Slug         string  `yaml:"slug" json:"slug"`
	Type         *string `yaml:"type,omitempty" json:"type,omitempty"`
}

type UserRole struct {
	User     User    `yaml:"user" json:"user"`
	Role     *string `yaml:"role,omitempty" json:"role,omitempty"`
	Approved *bool   `yaml:"approved,omitempty" json:"approved,omitempty"`
	Status   *string `yaml:"status,omitempty" json:"status,omitempty"`
}
