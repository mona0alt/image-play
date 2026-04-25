package user

import "context"

type User struct {
	ID        int64
	OpenID    string
	Balance   float64
	FreeQuota int
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

func (s *Service) GetOrCreateByMockCode(ctx context.Context, code string) (*User, bool, error) {
	openID := "mock-openid-" + code

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
