package commandgraphql

import (
	"context"
	domain "cqrs-es-example-go/pkg/command/domain/errors"
	"cqrs-es-example-go/pkg/command/processor"
	"github.com/99designs/gqlgen/graphql"
	esa "github.com/j5ik2o/event-store-adapter-go"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func validationErrorHandling(ctx context.Context, errorList []error) {
	for _, err := range errorList {
		graphql.AddError(ctx, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: err.Error(),
			Extensions: map[string]interface{}{
				"code": "400",
			},
		})
	}
}

func errorHandling(ctx context.Context, err error) {
	switch e := err.(type) {
	case *processor.RepositoryError:
		var optimisticLockError esa.OptimisticLockError
		if errors.As(e.Cause, &optimisticLockError) {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: e.Error(),
				Extensions: map[string]interface{}{
					"code":  "409",
					"cause": optimisticLockError.Error(),
				},
			})
		} else {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: e.Error(),
				Extensions: map[string]interface{}{
					"code": "500",
				},
			})
		}
	case *processor.NotFoundError:
		graphql.AddError(ctx, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: e.Error(),
			Extensions: map[string]interface{}{
				"code": "404",
			},
		})
	case *processor.DomainLogicError:
		var groupChatError domain.GroupChatError
		if errors.As(e.Cause, &groupChatError) {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: e.Error(),
				Extensions: map[string]interface{}{
					"code":  "422",
					"cause": groupChatError.Error(),
				},
			})
		} else {
			graphql.AddError(ctx, &gqlerror.Error{
				Path:    graphql.GetPath(ctx),
				Message: e.Error(),
				Extensions: map[string]interface{}{
					"code":  "500",
					"cause": e.Cause.Error(),
				},
			})
		}
	default:
		graphql.AddError(ctx, &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: err.Error(),
			Extensions: map[string]interface{}{
				"code": "500",
			},
		})
	}
}
