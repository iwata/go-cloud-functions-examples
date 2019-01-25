package function

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/functions/metadata"
	slack "github.com/ashwanthkumar/slack-go-webhook"
	"github.com/pkg/errors"
	"google.golang.org/api/cloudbuild/v1"
)

const SlackWebhookURL = "[SLACK_WEBHOOK]"

var (
	projectID string
	resource  string
	// Skip if the current status is not in the status list.
	// Add additional statues to list if you'd like:
	// QUEUED, WORKING, SUCCESS, FAILURE,
	// INTERNAL_ERROR, TIMEOUT, CANCELLED
	status = map[string]bool{
		"SUCCESS":        true,
		"FAILURE":        true,
		"INTERNAL_ERROR": true,
		"TIMEOUT":        true,
	}
)

func init() {
	projectID = os.Getenv("GCP_PROJECT")
	resource = fmt.Sprintf("projects/%s/topics/cloud-builds", projectID)
}

type PubSubMessage struct {
	Data string `json:"data"`
}

// Subscribe is the main function called by Cloud Functions.
func Subscribe(ctx context.Context, m PubSubMessage) error {
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get metadata")
	}
	if meta.Resource.Name != resource {
		fmt.Printf("%s is not wathing resource\n", meta.Resource.Name)
		return nil
	}

	build, err := eventToBuild(m.Data)
	if err != nil {
		return errors.Wrap(err, "Failed to decode event data")
	}

	if _, ok := status[build.Status]; !ok {
		fmt.Printf("%s status is skipped\n", build.Status)
		return nil
	}

	// Send message to Slack.
	message := createSlackMessage(build)
	errs := slack.Send(SlackWebhookURL, "", message)
	if len(errs) > 0 {
		return errors.Errorf("Failed to send a message to Slack: %s", errs)
	}

	return nil
}

// eventToBuild transforms pubsub event message to a Build struct.
func eventToBuild(data string) (*cloudbuild.Build, error) {
	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode base64 data")
	}

	build := cloudbuild.Build{}
	err = json.Unmarshal(d, &build)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode to JSON")
	}
	return &build, nil
}

// createSlackMessage creates a message from a build object.
func createSlackMessage(build *cloudbuild.Build) slack.Payload {
	title := "Build Logs"
	a := slack.Attachment{
		Title:     &title,
		TitleLink: &build.LogUrl,
	}
	a.AddField(slack.Field{
		Title: "Status",
		Value: build.Status,
	})
	p := slack.Payload{
		Text:        fmt.Sprintf("Build `%s`", build.Id),
		Markdown:    true,
		Attachments: []slack.Attachment{a},
	}
	return p
}
