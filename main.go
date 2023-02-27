package main

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

func main() {
	SetupGlobals()

	ctx := context.Background()
	collection := NewCollection(ctx, "m_test")

	path := 2
	collection.Delete(ctx, path)
	data := map[string]any{"path": path}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			collection.AddIfNotExists(ctx, data, 0, strconv.Itoa(i))
			wg.Done()
		}(i)
	}
	wg.Wait()
}

type Collection struct {
	client     *firestore.Client
	collection *firestore.CollectionRef
	log        *zap.Logger
}

func NewCollection(ctx context.Context, collectionName string) *Collection {
	log := L.With(zap.String("name", collectionName))

	var err error
	client, err := firestore.NewClient(ctx, FirestoreProjectID, option.WithCredentialsJSON(FirestoreCredentials))
	if err != nil {
		log.Error("Unable create Firestore client")
		return nil
	}

	collection := client.Collection(collectionName)

	return &Collection{
		log:        log,
		collection: collection,
		client:     client}
}

func (c Collection) Add(ctx context.Context, data map[string]any) error {
	_, _, err := c.collection.Add(ctx, data)
	if err != nil {
		c.log.Error("Unalbe to add document to collection", zap.Any("data", data))
	}
	return err
}

func (c Collection) Delete(ctx context.Context, path int) error {
	docIter := c.collection.Where("path", "==", path).Limit(1).Documents(ctx)

	d, err := docIter.Next()
	if err != nil {
		return err
	}
	_, err = d.Ref.Delete(ctx)
	return err
}

func (c Collection) AddIfNotExists(ctx context.Context, data map[string]any, sleep time.Duration, id string) error {
	log := c.log.With(zap.String("id", id))
	log.Debug("entering")
	q := c.collection.Where("path", "==", data["path"])
	err := c.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docs, err := tx.Documents(q).GetAll()
		if err != nil {
			return err
		}
		time.Sleep(sleep)
		if len(docs) > 0 {
			return errors.New("already exists")
		}

		return tx.Create(c.collection.Doc(uuid.New().String()), data)
	})
	if err != nil {
		log.Error("Tx error", zap.Error(err))
	}
	return err
}
