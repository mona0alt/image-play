package billing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
)

var ErrGenerationAlreadyCharged = errors.New("generation already charged")

// User represents a user with billing attributes.
type User struct {
	ID        int64
	Balance   float64
	FreeQuota int
}

// Transaction represents a billing transaction.
type Transaction struct {
	ID            int64
	UserID        int64
	GenerationID  int64
	Type          string
	Amount        float64
	BalanceBefore float64
	BalanceAfter  float64
}

// Order represents a payment order.
type Order struct {
	ID          int64
	UserID      int64
	OrderNo     string
	PackageCode string
	Amount      float64
	Status      string
	WxOrderNo   string
}

// Repository defines persistence operations for billing.
type Repository interface {
	GetUser(ctx context.Context, id int64) (*User, error)
	CreateUser(ctx context.Context, id int64, balance float64) error
	DeductBalance(ctx context.Context, userID int64, amount float64) error
	AddBalance(ctx context.Context, userID int64, amount float64) error
	DecrementFreeQuota(ctx context.Context, userID int64) error
	CreateTransaction(ctx context.Context, t *Transaction) error
	GetTransactionByGeneration(ctx context.Context, generationID int64) (*Transaction, error)
	CreateOrder(ctx context.Context, o *Order) error
	GetOrderByWxOrderNo(ctx context.Context, wxOrderNo string) (*Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status string) error
	AtomicCharge(ctx context.Context, userID, generationID int64, amount float64, useFreeQuota bool) error
	AtomicRecharge(ctx context.Context, userID, orderID int64, amount float64) error
}

// Service provides billing operations.
type Service struct {
	repo Repository
}

// NewService creates a new billing service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ChargeGeneration deducts from the user's free_quota first, then balance.
// If the generation was already charged, it returns ErrGenerationAlreadyCharged.
func (s *Service) ChargeGeneration(ctx context.Context, userID, generationID int64) error {
	// Fast-path idempotency check (optional, avoids DB tx overhead)
	existing, err := s.repo.GetTransactionByGeneration(ctx, generationID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrGenerationAlreadyCharged
	}

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	cost := 10.00
	useFreeQuota := user.FreeQuota > 0

	if err := s.repo.AtomicCharge(ctx, userID, generationID, cost, useFreeQuota); err != nil {
		if errors.Is(err, ErrGenerationAlreadyCharged) {
			return ErrGenerationAlreadyCharged
		}
		return err
	}
	return nil
}

// PackageDTO represents a purchasable package.
type PackageDTO struct {
	Code  string `json:"code"`
	Title string `json:"title"`
	Price string `json:"price"`
	Count int    `json:"count"`
}

// GetPackages returns available packages.
func (s *Service) GetPackages(ctx context.Context) ([]PackageDTO, error) {
	return []PackageDTO{
		{Code: "basic", Title: "基础套餐", Price: "10.00", Count: 1},
		{Code: "pro", Title: "专业套餐", Price: "50.00", Count: 6},
		{Code: "unlimited", Title: "无限套餐", Price: "100.00", Count: 15},
	}, nil
}

// OrderDTO represents order creation response.
type OrderDTO struct {
	OrderNo     string `json:"order_no"`
	PackageCode string `json:"package_code"`
	Amount      string `json:"amount"`
	PrepayID    string `json:"prepay_id"`
}

// CreateOrder creates a payment order.
func (s *Service) CreateOrder(ctx context.Context, userID int64, packageCode string) (*OrderDTO, error) {
	packages, err := s.GetPackages(ctx)
	if err != nil {
		return nil, err
	}
	var selected *PackageDTO
	for _, p := range packages {
		if p.Code == packageCode {
			selected = &p
			break
		}
	}
	if selected == nil {
		return nil, errors.New("invalid package code")
	}

	orderNo := generateOrderNo()
	price, err := parsePrice(packages, packageCode)
	if err != nil {
		return nil, err
	}
	o := &Order{
		UserID:      userID,
		OrderNo:     orderNo,
		PackageCode: packageCode,
		Amount:      price,
		Status:      "pending",
		WxOrderNo:   "",
	}
	if err := s.repo.CreateOrder(ctx, o); err != nil {
		return nil, err
	}

	return &OrderDTO{
		OrderNo:     orderNo,
		PackageCode: packageCode,
		Amount:      selected.Price,
		PrepayID:    "mock-prepay-id-" + orderNo,
	}, nil
}

// HandlePaymentCallback processes WeChat pay callback (mock for MVP).
func (s *Service) HandlePaymentCallback(ctx context.Context, wxOrderNo string) error {
	order, err := s.repo.GetOrderByWxOrderNo(ctx, wxOrderNo)
	if err != nil {
		return err
	}
	if order.Status == "paid" {
		return nil // idempotent
	}
	return s.repo.AtomicRecharge(ctx, order.UserID, order.ID, order.Amount)
}

func generateOrderNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "ORD" + hex.EncodeToString(b)
}

func parsePrice(packages []PackageDTO, code string) (float64, error) {
	for _, p := range packages {
		if p.Code == code {
			return strconv.ParseFloat(p.Price, 64)
		}
	}
	return 0, errors.New("package not found")
}
