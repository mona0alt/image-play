package user

import (
	"context"
	"math/rand"
	"strconv"

	"image-play/internal/infrastructure/wechat"
)

type User struct {
	ID        int64
	OpenID    string
	Balance   float64
	FreeQuota int
	Nickname  string
	AvatarURL string
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByOpenID(ctx context.Context, openID string) (*User, error)
	Create(ctx context.Context, user *User) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type WechatClient interface {
	Code2Session(ctx context.Context, code string) (*wechat.Code2SessionResponse, error)
}

func (s *Service) GetOrCreateByWxCode(ctx context.Context, code string, wxClient WechatClient) (*User, bool, error) {
	session, err := wxClient.Code2Session(ctx, code)
	if err != nil {
		return nil, false, err
	}
	openID := session.OpenID

	existing, err := s.repo.GetByOpenID(ctx, openID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	account := &User{
		OpenID:    openID,
		Balance:   0,
		FreeQuota: 3,
		Nickname:  "创作者" + strconv.Itoa(rand.Intn(900)+100),
	}
	if err := s.repo.Create(ctx, account); err != nil {
		existing, getErr := s.repo.GetByOpenID(ctx, openID)
		if getErr != nil {
			return nil, false, getErr
		}
		if existing != nil {
			return existing, false, nil
		}
		return nil, false, err
	}

	return account, true, nil
}
