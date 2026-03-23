package backlogtest

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/pkg/backlog"
)

func NewIndex(t *testing.T, opts ...backlog.IndexOption) *backlog.Index {
	t.Helper()

	ind, err := backlog.New(t.Context(), opts...)
	require.NoError(t, err)

	return ind
}
