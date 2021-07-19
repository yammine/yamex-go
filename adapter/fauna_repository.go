package adapter

import (
	"fmt"
	"hash/fnv"

	"github.com/fauna/faunadb-go/v4/faunadb"

	"github.com/yammine/yamex-go/app"
	"github.com/yammine/yamex-go/domain"
)

type FaunaClient interface {
	Query(expr faunadb.Expr, configs ...faunadb.QueryConfig) (value faunadb.Value, err error)
}

type FaunaRepository struct {
	client FaunaClient
}

func NewFaunaRepository(client FaunaClient) *FaunaRepository {
	return &FaunaRepository{client: client}
}

func (f FaunaRepository) GetOrCreateUser(id string) (*domain.User, error) {
	hashed := numericalHash(id)

	userRef := faunadb.Ref(faunadb.Collection(UsersCollection), hashed)
	value, err := f.client.Query(
		faunadb.If(
			faunadb.Exists(userRef),
			faunadb.Get(userRef),
			faunadb.Create(userRef, faunadb.Obj{"data": faunadb.Obj{"slack_user_id": id}}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("get or create user query: %w", err)
	}

	user := &domain.User{}
	value.At(faunadb.ObjKey("data")).Get(user)
	// Decode `ref` into user.ID
	var ref faunadb.RefV
	value.At(faunadb.ObjKey("ref")).Get(&ref)
	user.ID = ref.ID

	return user, nil
}

var _ app.Repository = (*FaunaRepository)(nil)

func numericalHash(id string) string {
	h := fnv.New64()
	h.Write([]byte(id))
	s := fmt.Sprint(h.Sum64())
	return s[0:18]
}
