package gcf

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildEvent_AvailableStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"SUCCESS", true},
		{"FAILURE", true},
		{"INTERNAL_ERROR", true},
		{"TIMEOUT", true},
		{"QUEUED", false},
		{"WORKING", false},
		{"CANCELLED", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Status: tt.status,
			}
			if got := e.AvailableStatus(); got != tt.want {
				t.Errorf("BuildEvent.AvailableStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEvent_SlackStatus(t *testing.T) {
	tests := []struct {
		status string
		want   SlackStatus
	}{
		{"SUCCESS", SlackStatus{Color: "#2aa24b", Icon: ":white_check_mark:"}},
		{"FAILURE", SlackStatus{Color: "#d50200", Icon: ":x:"}},
		{"INTERNAL_ERROR", SlackStatus{Color: "#d50200", Icon: ":sos:"}},
		{"TIMEOUT", SlackStatus{Color: "#de9d2e", Icon: ":sos:"}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Status: tt.status,
			}
			got := e.SlackStatus()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("BuildEvent.SlackStatus() = %v, want %v, differs: (-got +want;\n%s)", got, tt.want, diff)
			}
		})
	}
}

func TestBuildEvent_HasSource(t *testing.T) {
	tests := []struct {
		name   string
		source *BuildSource
		want   bool
	}{
		{"Has an available source", &BuildSource{&BuildRepoSource{BranchName: "master"}}, true},
		{"Has no available sources", nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Source: tt.source,
			}
			if got := e.HasSource(); got != tt.want {
				t.Errorf("BuildEvent.HasSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEvent_IsSuccess(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"SUCCESS", true},
		{"FAILURE", false},
		{"INTERNAL_ERROR", false},
		{"TIMEOUT", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.status, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Status: tt.status,
			}
			if got := e.IsSuccess(); got != tt.want {
				t.Errorf("BuildEvent.IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEvent_Branch(t *testing.T) {
	tests := []struct {
		name   string
		sub    *BuildSubstitutions
		source *BuildSource
		want   RepositoryBranch
	}{
		{
			"When using Cloud Build Github App",
			&BuildSubstitutions{BranchName: "master"},
			&BuildSource{&BuildRepoSource{BranchName: "dev"}},
			RepositoryBranch("master"),
		},
		{
			"Include substitutions field, but don't have branch name",
			&BuildSubstitutions{},
			&BuildSource{&BuildRepoSource{BranchName: "dev"}},
			RepositoryBranch("dev"),
		},
		{
			"Not include any substitutions",
			nil,
			&BuildSource{&BuildRepoSource{BranchName: "dev"}},
			RepositoryBranch("dev"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Source:        tt.source,
				Substitutions: tt.sub,
			}
			if got := e.Branch(); got != tt.want {
				t.Errorf("BuildEvent.Branch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEvent_IsDeploy(t *testing.T) {
	tests := []struct {
		name string
		tags *BuildTags
		want bool
	}{
		{"Don't have any tags", nil, false},
		{"Have some tags, but not includ deploy's one", &BuildTags{"deploy-unknown"}, false},
		{"Include 'deploy-default-service'", &BuildTags{"deploy-default-service", "another-tag"}, true},
		{"Include 'deploy-admin-service'", &BuildTags{"another-tag", "deploy-admin-service"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Tags: tt.tags,
			}
			if got := e.IsDeploy(); got != tt.want {
				t.Errorf("BuildEvent.IsDeploy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEvent_AppURLs(t *testing.T) {
	tests := []struct {
		name   string
		tags   *BuildTags
		branch string
		args   *SlackConfig
		want   []AppURL
	}{
		{
			"For deploying default service with master branch",
			&BuildTags{"deploy-default-service"},
			"master",
			&SlackConfig{ProjectID: "nomos-sms"},
			[]AppURL{
				{Title: "SMS URL", URL: "https://nomos-sms.appspot.com"},
				{Title: "SMS Career URL", URL: "https://smsc-dot-nomos-sms.appspot.com"},
			},
		},
		{
			"For deploying default service with dev branch",
			&BuildTags{"deploy-default-service"},
			"dev",
			&SlackConfig{ProjectID: "nomos-sms"},
			[]AppURL{
				{Title: "SMS URL", URL: "https://dev-dot-nomos-sms.appspot.com"},
				{Title: "SMS Career URL", URL: "https://dev-dot-smsc-dot-nomos-sms.appspot.com"},
			},
		},
		{
			"For deploying admin service with master branch",
			&BuildTags{"deploy-admin-service"},
			"master",
			&SlackConfig{ProjectID: "nomos-sms"},
			[]AppURL{
				{Title: "Admin URL", URL: "https://admin-dot-nomos-sms.appspot.com"},
			},
		},
		{
			"For deploying admin service with dev branch",
			&BuildTags{"deploy-admin-service"},
			"dev",
			&SlackConfig{ProjectID: "nomos-sms"},
			[]AppURL{
				{Title: "Admin URL", URL: "https://dev-dot-admin-dot-nomos-sms.appspot.com"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := BuildEvent{
				Source: &BuildSource{
					RepoSource: &BuildRepoSource{
						BranchName: tt.branch,
					},
				},
				Tags: tt.tags,
			}
			got := e.AppURLs(tt.args)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("BuildEvent.AppURLs() = %v, want %v, differs: (-got +want;\n%s)", got, tt.want, diff)
			}
		})
	}
}

func TestRepositoryBranch_ToVersion(t *testing.T) {
	tests := []struct {
		name string
		b    RepositoryBranch
		want string
	}{
		{
			"Over max length of version name",
			RepositoryBranch(strings.Repeat("a", 41)),
			"10c4da18575c092b486f8ab96c01c02f",
		},
		{
			"Replace invalid characters to hyphen",
			RepositoryBranch("@dependabot/go-lang/cloud_build.go"),
			"-dependabot-go-lang-cloud-build-go",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.b.ToVersion(); got != tt.want {
				t.Errorf("RepositoryBranch.ToVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
