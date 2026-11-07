package querygraphql

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestReadModelError(t *testing.T) {
	ctx := context.Background()

	notFound := readModelError(ctx, "GroupChat", sql.ErrNoRows).(*gqlerror.Error)
	require.Equal(t, "404", notFound.Extensions["code"])
	require.Equal(t, "GroupChat not found.", notFound.Message)

	internal := readModelError(ctx, "GroupChat", errors.New("dial tcp 10.0.0.1:3306: connect: connection refused")).(*gqlerror.Error)
	require.Equal(t, "500", internal.Extensions["code"])
	require.Equal(t, "internal server error", internal.Message)
	require.NotContains(t, internal.Message, "10.0.0.1")
}
