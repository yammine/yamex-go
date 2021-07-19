package adapter

import (
	"fmt"

	f "github.com/fauna/faunadb-go/v4/faunadb"
)

const (
	UsersCollection    = "users"
	UsersBySlackUserID = "users_by_slack_user_id"
)

func InitFaunaDatabase(client *f.FaunaClient) error {
	var err error
	if err = CreateCollections(client); err != nil {
		return fmt.Errorf("creating collections: %w", err)
	}
	if err = CreateIndices(client); err != nil {
		return fmt.Errorf("creating indices: %w", err)
	}

	return nil
}

func CreateCollections(client *f.FaunaClient) error {
	_, err := client.Query(
		f.Map(
			f.Arr{UsersCollection},
			f.Lambda("col", f.If(
				f.Exists(f.Collection(f.Var("col"))),
				f.Null(),
				f.CreateCollection(f.Obj{"name": f.Var("col")}),
			)),
		),
	)

	return err
}

func CreateIndices(client *f.FaunaClient) error {
	_, err := client.Query(f.If(
		f.Exists(f.Index(UsersBySlackUserID)),
		f.Null(),
		f.CreateIndex(
			f.Obj{
				"name":       UsersBySlackUserID,
				"source":     f.Collection(UsersCollection),
				"terms":      f.Arr{f.Obj{"field": f.Arr{"data", "slack_user_id"}}},
				"unique":     true,
				"serialized": true,
			},
		),
	),
	)

	return err
}
