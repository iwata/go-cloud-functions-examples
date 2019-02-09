package gcf

import "testing"

func TestSlackConfig(t *testing.T) {
	projectID := "test-project"
	webhook := "hook"
	config := &SlackConfig{
		ProjectID:    projectID,
		SlackWebhook: webhook,
	}

	got := config.DefaultDomain()
	want := "test-project.appspot.com"
	if got != want {
		t.Errorf("SlackConfig.DefaultDomain() returns %s, but want %s", got, want)
	}

	got = config.CareerDomain()
	want = "smsc-dot-test-project.appspot.com"
	if got != want {
		t.Errorf("SlackConfig.CareerDomain() returns %s, but want %s", got, want)
	}

	got = config.AdminDomain()
	want = "admin-dot-test-project.appspot.com"
	if got != want {
		t.Errorf("SlackConfig.AdminDomain() returns %s, but want %s", got, want)
	}

	got = config.SlackWebhookURL()
	want = "https://hooks.slack.com/services/hook"
	if got != want {
		t.Errorf("SlackConfig.SlackWebhookURL() returns %s, but want %s", got, want)
	}

	got = config.WatchingResource()
	want = "projects/test-project/topics/cloud-builds"
	if got != want {
		t.Errorf("SlackConfig.SlackWebhookURL() returns %s, but want %s", got, want)
	}
}

func TestFirestoreConfig(t *testing.T) {
	projectID := "test-project"
	config := &FirestoreConfig{ProjectID: projectID}

	if got, want := config.StorageURIPrefix(), "gs://test-project-backup-firestore"; got != want {
		t.Errorf("FirestoreConfig.StorageURIPrefix() returns %s, but want %s", got, want)
	}
	if got, want := config.DatabaseName(), "projects/test-project/databases/(default)"; got != want {
		t.Errorf("FirestoreConfig.DatabaseName() returns %s, but want %s", got, want)
	}
}
