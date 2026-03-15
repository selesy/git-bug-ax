package issuetest

import (
	"testing"

	"github.com/git-bug/git-bug/commands/bug/testenv"
	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/git-bug/git-bug/entity"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-ax/pkg/issue"
)

func NewEnv(t *testing.T) (*execenv.Env, entity.Id) {
	return testenv.NewTestEnvAndUser(t)
}

// func NewEnvWithIssues(t *testing.T) (*execenv.)

func NewTestIssue(t *testing.T, opts ...issue.Option) (*execenv.Env, *issue.Issue) {
	t.Helper()

	env, _ := testenv.NewTestEnvAndUser(t)
	user, err := env.Backend.GetUserIdentity()
	require.NoError(t, err)

	iss, err := issue.Create(env, opts...)
	require.NoError(t, err)
	require.NoError(t, iss.Commit(user))

	return env, iss
}
