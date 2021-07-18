package collections

import f "github.com/fauna/faunadb-go/v4/faunadb"

type collectionParams struct {
	Name string `fauna:"name"`
}

func CreateCollections(client *f.FaunaClient) error {
	_, err := client.Query(f.CreateCollection(collectionParams{Name: "users"}))

	return err
}
