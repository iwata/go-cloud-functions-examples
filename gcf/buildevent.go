package gcf

//nolint[:gosec]
import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

const (
	TagDeployDefault = "deploy-default-service"
	TagDeployAdmin   = "deploy-admin-service"
	RepositoryURL    = "https://github.com/bm-sms/nomos"
	MaxVersionLength = 40
)

type SlackStatus struct {
	Color string
	Icon  string
}

var statusMap = map[string]SlackStatus{
	"SUCCESS":        {Color: "#2aa24b", Icon: ":white_check_mark:"},
	"FAILURE":        {Color: "#d50200", Icon: ":x:"},
	"INTERNAL_ERROR": {Color: "#d50200", Icon: ":sos:"},
	"TIMEOUT":        {Color: "#de9d2e", Icon: ":sos:"},
}

type BuildEvent struct {
	ID             string              `json:"id"`
	ProjectID      string              `json:"projectId"`
	Status         string              `json:"status"`
	Source         *BuildSource        `json:"source"`
	CreateTime     time.Time           `json:"createTime"`
	StartTime      time.Time           `json:"startTime"`
	FinishTime     time.Time           `json:"finishTime"`
	Timeout        string              `json:"timeout"`
	LogsBucket     string              `json:"logsBucket"`
	BuildTriggerID string              `json:"buildTriggerId"`
	LogURL         string              `json:"logUrl"`
	Tags           *BuildTags          `json:"tags"`
	Substitutions  *BuildSubstitutions `json:"substitutions"`
}

type BuildSubstitutions struct {
	BranchName string `json:"BRANCH_NAME"`
}

type BuildSource struct {
	RepoSource *BuildRepoSource
}

type BuildRepoSource struct {
	ProjectID  string `json:"projectId"`
	RepoName   string `json:"repoName"`
	BranchName string `json:"branchName"`
}

type BuildTags []string

type AppURL struct {
	Title string
	URL   string
}

type RepositoryBranch string

func (e BuildEvent) AvailableStatus() bool {
	_, ok := statusMap[e.Status]
	return ok
}

func (e BuildEvent) SlackStatus() SlackStatus {
	return statusMap[e.Status]
}

func (e BuildEvent) HasSource() bool {
	return e.Source != nil
}

func (e BuildEvent) IsSuccess() bool {
	return e.Status == "SUCCESS"
}

func (e BuildEvent) Branch() RepositoryBranch {
	var br string
	// Cloud Build Github App has branch name in substitutions
	if e.Substitutions != nil && e.Substitutions.BranchName != "" {
		br = e.Substitutions.BranchName
	} else {
		br = e.Source.RepoSource.BranchName
	}

	return RepositoryBranch(br)
}

func (e BuildEvent) IsDeploy() bool {
	return e.Tags != nil && (e.Tags.includedTag(TagDeployDefault) || e.Tags.includedTag(TagDeployAdmin))
}

func (e BuildEvent) AppURLs(c *SlackConfig) []AppURL {
	if e.Tags.includedTag(TagDeployDefault) {
		return []AppURL{
			{Title: "SMS URL", URL: e.defaultURL(c)},
			{Title: "SMS Career URL", URL: e.careerURL(c)},
		}
	}

	return []AppURL{
		{Title: "Admin URL", URL: e.adminURL(c)},
	}
}

func (e BuildEvent) defaultURL(c *SlackConfig) string {
	return e.domainToURL(c.DefaultDomain())
}

func (e BuildEvent) careerURL(c *SlackConfig) string {
	return e.domainToURL(c.CareerDomain())
}

func (e BuildEvent) adminURL(c *SlackConfig) string {
	return e.domainToURL(c.AdminDomain())
}

func (e BuildEvent) domainToURL(domain string) string {
	if e.Branch().isMaster() {
		return fmt.Sprintf("https://%s", domain)
	}
	return fmt.Sprintf("https://%s-dot-%s", e.Branch().ToVersion(), domain)
}

func (b RepositoryBranch) URL() string {
	return fmt.Sprintf("%s/tree/%s", RepositoryURL, b)
}

//nolint[:gosec]
func (b RepositoryBranch) ToVersion() string {
	r := strings.NewReplacer("/", "-", ".", "-", "@", "-", "_", "-")
	v := r.Replace(strings.ToLower(string(b)))
	if len(v) < MaxVersionLength {
		return v
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(v)))
}

func (b RepositoryBranch) isMaster() bool {
	return string(b) == "master"
}

func (tags *BuildTags) includedTag(tag string) bool {
	for _, t := range []string(*tags) {
		if t == tag {
			return true
		}
	}
	return false
}
