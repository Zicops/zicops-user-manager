package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-user-manager/graph/generated"
	"github.com/zicops/zicops-user-manager/graph/model"
	"github.com/zicops/zicops-user-manager/handlers"
)

func (r *mutationResolver) RegisterUsers(ctx context.Context, input []*model.UserInput) ([]*model.User, error) {
	result, err := handlers.RegisterUsers(ctx, input)
	if err != nil {
		log.Errorf("Error registering users: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) InviteUsers(ctx context.Context, emails []string) (*bool, error) {
	result, err := handlers.InviteUsers(ctx, emails)
	if err != nil {
		log.Errorf("Error inviting users: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	result, err := handlers.UpdateUser(ctx, input)
	if err != nil {
		log.Errorf("Error updating user: %v", err)
		return nil, err
	}
	return result, nil
}

func (r *queryResolver) Todos(ctx context.Context) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
