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
	// Other setup functions here
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
