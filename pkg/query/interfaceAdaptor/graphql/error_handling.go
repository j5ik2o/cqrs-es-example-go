package querygraphql

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func errorHandling(ctx context.Context, err error) {
	graphql.AddError(ctx, &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": "500",
		},
	})
}
