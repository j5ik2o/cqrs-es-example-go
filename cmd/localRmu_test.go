package cmd

import (
	"errors"
	"testing"

	streamtypes "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/stretchr/testify/require"
)

func TestStreamErrorClassification(t *testing.T) {
	require.True(t, isExpiredIterator(&streamtypes.ExpiredIteratorException{}))
	require.False(t, isExpiredIterator(&streamtypes.TrimmedDataAccessException{}))
	require.False(t, isExpiredIterator(errors.New("other")))

	require.True(t, isTrimmedData(&streamtypes.TrimmedDataAccessException{}))
	require.False(t, isTrimmedData(&streamtypes.ExpiredIteratorException{}))
	require.False(t, isTrimmedData(errors.New("other")))

	// The SDK errors are wrapped by the operation layer; As must still see them.
	require.True(t, isTrimmedData(fmtWrap(&streamtypes.TrimmedDataAccessException{})))
	require.True(t, isExpiredIterator(fmtWrap(&streamtypes.ExpiredIteratorException{})))
}

func fmtWrap(err error) error {
	return errors.Join(errors.New("operation failed"), err)
}
