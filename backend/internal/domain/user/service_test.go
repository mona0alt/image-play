package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	nextID        int64
	usersByID     map[int64]*User
	usersByOpenID map[string]*User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		nextID:        1,
		usersByID:     make(map[int64]*User),
		usersByOpenID: make(map[string]*User),
	}
}

func (r *mockUserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}

	cloned := *user
	return &cloned, nil
}

func (r *mockUserRepo) GetByOpenID(_ context.Context, openID string) (*User, error) {
	user, ok := r.usersByOpenID[openID]
	if !ok {
		return nil, nil
	}

	cloned := *user
	return &cloned, nil
}

func (r *mockUserRepo) Create(_ context.Context, user *User) error {
	user.ID = r.nextID
	r.nextID++

	cloned := *user
	r.usersByID[user.ID] = &cloned
	r.usersByOpenID[user.OpenID] = &cloned
	return nil
}

func TestGetOrCreateByMockCodeCreatesUserOnFirstLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	user, created, err := svc.GetOrCreateByMockCode(context.Background(), "wx-code-1")

	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "mock-openid-wx-code-1", user.OpenID)
	require.Equal(t, 3, user.FreeQuota)
}

func TestGetOrCreateByMockCodeReusesExistingUser(t *testing.T) {
	repo := newMockUserRepo()
	repo.usersByOpenID["mock-openid-wx-code-1"] = &User{
		ID:        42,
		OpenID:    "mock-openid-wx-code-1",
		Balance:   0,
		FreeQuota: 2,
	}
	svc := NewService(repo)

	user, created, err := svc.GetOrCreateByMockCode(context.Background(), "wx-code-1")

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(42), user.ID)
	require.Equal(t, 2, user.FreeQuota)
}
