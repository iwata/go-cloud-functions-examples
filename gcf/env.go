package gcf

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/tenntenn/sync/try"
)

type SlackConfig struct {
	ProjectID    string `envconfig:"gcp_project"`
	SlackWebhook string `envconfig:"slack_webhook"`
}

type FirestoreConfig struct {
	ProjectID string `envconfig:"gcp_project"`
}

var (
	slackConfig         SlackConfig
	onceSlackConfig     try.Once
	firestoreConfig     FirestoreConfig
	onceFirestoreConfig try.Once
)

func getSlackConfig() (*SlackConfig, error) {
	err := onceSlackConfig.Try(func() error {
		return envconfig.Process("", &slackConfig)
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &slackConfig, nil
}

func (c *SlackConfig) DefaultDomain() string {
	return fmt.Sprintf("%s.appspot.com", c.ProjectID)
}

func (c *SlackConfig) CareerDomain() string {
	return fmt.Sprintf("smsc-dot-%s", c.DefaultDomain())
}

func (c *SlackConfig) AdminDomain() string {
	return fmt.Sprintf("admin-dot-%s", c.DefaultDomain())
}

func (c *SlackConfig) SlackWebhookURL() string {
	return fmt.Sprintf("https://hooks.slack.com/services/%s", c.SlackWebhook)
}

func (c *SlackConfig) WatchingResource() string {
	return fmt.Sprintf("projects/%s/topics/cloud-builds", c.ProjectID)
}

func getFirestoreConfig() (*FirestoreConfig, error) {
	err := onceFirestoreConfig.Try(func() error {
		return envconfig.Process("", &firestoreConfig)
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &firestoreConfig, nil
}

func (c *FirestoreConfig) StorageURIPrefix() string {
	return fmt.Sprintf("gs://%s-backup-firestore", c.ProjectID)
}

func (c *FirestoreConfig) DatabaseName() string {
	return fmt.Sprintf("projects/%s/databases/(default)", c.ProjectID)
}
