package querygraphql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func readModelError(ctx context.Context, resource string, err error) error {
	code := "500"
	message := err.Error()
	if errors.Is(err, sql.ErrNoRows) {
		code = "404"
		message = resource + " not found."
	}
	return &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}
