package backlogtest

import (
	"testing"

	"github.com/git-bug/git-bug/commands/bug/testenv"
	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/pkg/backlog"
)

func NewTestIssue(t *testing.T, opts ...backlog.IssueOption) (*execenv.Env, *backlog.Issue) {
	t.Helper()

	env, _ := testenv.NewTestEnvAndUser(t)
	user, err := env.Backend.GetUserIdentity()
	require.NoError(t, err)

	iss, err := backlog.Create(env, opts...)
	require.NoError(t, err)
	require.NoError(t, iss.Commit(user))

	return env, iss
}
