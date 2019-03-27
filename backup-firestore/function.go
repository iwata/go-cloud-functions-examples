package function

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/api/firestore/v1"
	"google.golang.org/api/option"
)

var projectID string

// PubSubMessage is a body data from Pub/Sub payload
type PubSubMessage struct {
	Data string `json:"data"`
}

func init() {
	projectID = os.Getenv("GCP_PROJECT")
}

// BackupFirestore is triggered by Cloud Functions
func BackupFirestore(ctx context.Context, m PubSubMessage) error {
	svc, err := firestore.NewService(ctx, option.WithScopes(firestore.DatastoreScope, firestore.CloudPlatformScope))
	if err != nil {
		return errors.Wrap(err, "Failed to create Firestore service")
	}

	req := &firestore.GoogleFirestoreAdminV1ExportDocumentsRequest{
		OutputUriPrefix: fmt.Sprintf("gs://%s-backup-firestore", projectID),
	}
	_, err = firestore.NewProjectsDatabasesService(svc).ExportDocuments(
		fmt.Sprintf("projects/%s/databases/(default)", projectID), req,
	).Context(ctx).Do()
	if err != nil {
		return errors.Wrap(err, "Failed to export Firestore")
	}

	fmt.Println("Backup Successfully")

	return nil
}
