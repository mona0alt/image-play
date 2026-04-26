package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"image-play/internal/infrastructure/wechat"
)

type mockWxClient struct {
	openID string
	err    error
}

func (m *mockWxClient) Code2Session(_ context.Context, _ string) (*wechat.Code2SessionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &wechat.Code2SessionResponse{
		OpenID:     m.openID,
		SessionKey: "session-key",
	}, nil
}

type mockUserRepo struct {
	nextID        int64
	usersByID     map[int64]*User
	usersByOpenID map[string]*User
	createErr     error
	onCreate      func(*User)
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
	if r.onCreate != nil {
		r.onCreate(user)
	}
	if r.createErr != nil {
		return r.createErr
	}

	user.ID = r.nextID
	r.nextID++

	cloned := *user
	r.usersByID[user.ID] = &cloned
	r.usersByOpenID[user.OpenID] = &cloned
	return nil
}

func (r *mockUserRepo) UpdateNickname(_ context.Context, id int64, nickname string) error {
	if user, ok := r.usersByID[id]; ok {
		user.Nickname = nickname
	}
	return nil
}

func TestGetOrCreateByWxCodeCreatesUserOnFirstLogin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1"}

	user, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-1", wx)

	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "wx-openid-1", user.OpenID)
	require.Equal(t, 3, user.FreeQuota)
	require.True(t, strings.HasPrefix(user.Nickname, "创作者"))
}

func TestGetOrCreateByWxCodeReusesExistingUser(t *testing.T) {
	repo := newMockUserRepo()
	repo.usersByOpenID["wx-openid-1"] = &User{
		ID:        42,
		OpenID:    "wx-openid-1",
		Balance:   0,
		FreeQuota: 2,
		Nickname:  "ExistingUser",
	}
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-1"}

	user, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-1", wx)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(42), user.ID)
	require.Equal(t, 2, user.FreeQuota)
	require.Equal(t, "ExistingUser", user.Nickname)
}

func TestGetOrCreateByWxCodeReusesUserWhenCreateLosesRace(t *testing.T) {
	repo := newMockUserRepo()
	repo.createErr = errors.New("duplicate key value violates unique constraint")
	repo.onCreate = func(user *User) {
		repo.usersByID[99] = &User{
			ID:        99,
			OpenID:    user.OpenID,
			Balance:   0,
			FreeQuota: 3,
			Nickname:  "RacedUser",
		}
		repo.usersByOpenID[user.OpenID] = repo.usersByID[99]
	}
	svc := NewService(repo)
	wx := &mockWxClient{openID: "wx-openid-race"}

	account, created, err := svc.GetOrCreateByWxCode(context.Background(), "wx-code-race", wx)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, int64(99), account.ID)
	require.Equal(t, "wx-openid-race", account.OpenID)
	require.Equal(t, 3, account.FreeQuota)
}

func TestGetOrCreateByWxCodeReturnsWechatError(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)
	wx := &mockWxClient{err: errors.New("wechat api error")}

	user, created, err := svc.GetOrCreateByWxCode(context.Background(), "bad-code", wx)

	require.Error(t, err)
	require.Nil(t, user)
	require.False(t, created)
}
