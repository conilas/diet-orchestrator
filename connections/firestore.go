package connections

import (
	firestore "cloud.google.com/go/firestore"
	"context"
	"log"
)

func CreateFirestoreConnection() *firestore.Client {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, "scheduler")
	if err != nil {
		log.Fatalln(err)
	}

	return client
}
