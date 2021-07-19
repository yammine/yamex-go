package adapter

import (
	"fmt"

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

func (f FaunaRepository) GetOrCreateUser(slackUserId string) (*domain.User, error) {
	result, err := f.client.Query(
		faunadb.Let().
			Bind("match_result", faunadb.MatchTerm(faunadb.Index(UsersBySlackUserID), slackUserId)).
			In(
				faunadb.If(
					faunadb.IsNonEmpty(faunadb.Var("match_result")),
					faunadb.Get(faunadb.Var("match_result")),
					faunadb.Create(faunadb.Collection(UsersCollection), faunadb.Obj{"data": faunadb.Obj{"slack_user_id": slackUserId}}),
				),
			),
	)
	if err != nil {
		return nil, fmt.Errorf("get or create user query: %w", err)
	}

	user := &domain.User{}
	result.At(faunadb.ObjKey("data")).Get(user)
	// Decode `ref` into user.ID
	var ref faunadb.RefV
	result.At(faunadb.ObjKey("ref")).Get(&ref)
	user.ID = ref.ID

	return user, nil
}

var _ app.Repository = (*FaunaRepository)(nil)
