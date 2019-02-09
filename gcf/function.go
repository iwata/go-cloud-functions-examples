package gcf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/functions/metadata"
	slack "github.com/ashwanthkumar/slack-go-webhook"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/firestore/v1"
)

type PubSubMessage struct {
	Data string `json:"data"`
}

const (
	Service = "Nomos"
)

func NotifySlack(ctx context.Context, m PubSubMessage) error {
	config, err := getSlackConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to get config about Slack")
	}

	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get metadata")
	}
	if meta.Resource.Name != config.WatchingResource() {
		fmt.Printf("%s is not wathing resource\n", meta.Resource.Name)
		return nil
	}

	d, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil {
		return errors.Wrap(err, "Failed to decode base64 data")
	}
	fmt.Printf("Build data: %s\n", string(d))

	build := BuildEvent{}
	err = json.Unmarshal(d, &build)
	if err != nil {
		return errors.Wrap(err, "Failed to decode to JSON")
	}

	if !build.HasSource() {
		fmt.Print("This event isn't source code build")
		return nil
	}

	if !build.AvailableStatus() {
		fmt.Printf("%s is non available status\n", build.Status)
		return nil
	}

	payload := createSlackPayload(build, config)
	errs := slack.Send(config.SlackWebhookURL(), "", payload)
	if len(errs) > 0 {
		return errors.Errorf("Failed to send a message to Slack: %s", errs)
	}
	fmt.Println("Sent a message to Slack")

	return nil
}

func createSlackPayload(b BuildEvent, c *SlackConfig) slack.Payload {
	title := "Build Logs"
	color := b.SlackStatus().Color
	a := slack.Attachment{
		Title:      &title,
		TitleLink:  &b.LogURL,
		Color:      &color,
		MarkdownIn: &[]string{"fields"},
	}
	a.AddField(slack.Field{
		Title: "status",
		Value: fmt.Sprintf("%s %s", b.SlackStatus().Icon, b.Status),
		Short: true,
	}).AddField(slack.Field{
		Title: "Branch",
		Value: fmt.Sprintf("<%s|%s>", b.Branch().URL(), b.Branch()),
		Short: true,
	})

	if b.IsSuccess() && b.IsDeploy() {
		urls := b.AppURLs(c)
		for _, u := range urls {
			a.AddField(slack.Field{
				Title: u.Title,
				Value: u.URL,
			})
		}
	}

	a.AddField(slack.Field{
		Title: "Tag",
		Value: []string(*b.Tags)[0],
	})

	p := slack.Payload{
		Username:    "Cloud Build",
		IconEmoji:   ":cloudbuild:",
		Text:        fmt.Sprintf("%s was built as %s", Service, b.ID),
		Markdown:    true,
		Attachments: []slack.Attachment{a},
	}

	return p
}

func BackupFirestore(ctx context.Context, m PubSubMessage) error {
	fmt.Printf("Message: %#v\n", m)

	config, err := getFirestoreConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to get config about Firestore")
	}

	client, err := google.DefaultClient(ctx,
		"https://www.googleapis.com/auth/datastore",
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return errors.Wrap(err, "Failed to create a Google client")
	}

	svc, err := firestore.New(client)
	if err != nil {
		return errors.Wrap(err, "Failed to create Firestore service")
	}

	req := &firestore.GoogleFirestoreAdminV1ExportDocumentsRequest{
		OutputUriPrefix: config.StorageURIPrefix(),
	}
	res, err := firestore.NewProjectsDatabasesService(svc).ExportDocuments(
		config.DatabaseName(), req,
	).Context(ctx).Do()
	if err != nil {
		return errors.Wrap(err, "Failed to export Firestore")
	}
	fmt.Printf("Successful Response: %#v\n", res)

	return nil
}
