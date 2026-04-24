package billing

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type mockRepo struct {
	mu           sync.Mutex
	users        map[int64]*User
	transactions []*Transaction
	orders       []*Order
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		users:        make(map[int64]*User),
		transactions: make([]*Transaction, 0),
		orders:       make([]*Order, 0),
	}
}

func (r *mockRepo) GetUser(_ context.Context, id int64) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.users[id]; ok {
		// Return a copy
		cp := *u
		return &cp, nil
	}
	return nil, errors.New("user not found")
}

func (r *mockRepo) CreateUser(_ context.Context, id int64, balance float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[id] = &User{ID: id, Balance: balance, FreeQuota: 0}
	return nil
}

func (r *mockRepo) DeductBalance(_ context.Context, userID int64, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	if u.Balance < amount {
		return errors.New("insufficient balance")
	}
	u.Balance -= amount
	return nil
}

func (r *mockRepo) AddBalance(_ context.Context, userID int64, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.Balance += amount
	return nil
}

func (r *mockRepo) DecrementFreeQuota(_ context.Context, userID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	if u.FreeQuota <= 0 {
		return errors.New("no free quota")
	}
	u.FreeQuota--
	return nil
}

func (r *mockRepo) CreateTransaction(_ context.Context, t *Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.ID = int64(len(r.transactions) + 1)
	r.transactions = append(r.transactions, t)
	return nil
}

func (r *mockRepo) GetTransactionByGeneration(_ context.Context, generationID int64) (*Transaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.transactions {
		if t.GenerationID == generationID {
			cp := *t
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *mockRepo) CreateOrder(_ context.Context, o *Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	o.ID = int64(len(r.orders) + 1)
	r.orders = append(r.orders, o)
	return nil
}

func (r *mockRepo) GetOrderByWxOrderNo(_ context.Context, wxOrderNo string) (*Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.orders {
		if o.WxOrderNo == wxOrderNo {
			cp := *o
			return &cp, nil
		}
	}
	return nil, errors.New("order not found")
}

func (r *mockRepo) UpdateOrderStatus(_ context.Context, id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.orders {
		if o.ID == id {
			o.Status = status
			return nil
		}
	}
	return errors.New("order not found")
}

func (r *mockRepo) AtomicCharge(_ context.Context, userID, generationID int64, amount float64, useFreeQuota bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.transactions {
		if t.GenerationID == generationID {
			return ErrGenerationAlreadyCharged
		}
	}
	u, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	balanceBefore := u.Balance
	transAmount := amount
	if useFreeQuota {
		if u.FreeQuota <= 0 {
			return errors.New("insufficient free quota")
		}
		u.FreeQuota--
		balanceBefore = 0
		transAmount = 0
	} else {
		if u.Balance < amount {
			return errors.New("insufficient balance")
		}
		u.Balance -= amount
	}
	r.transactions = append(r.transactions, &Transaction{
		ID:            int64(len(r.transactions) + 1),
		UserID:        userID,
		GenerationID:  generationID,
		Type:          "generation_charge",
		Amount:        transAmount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceBefore - transAmount,
	})
	return nil
}

func (r *mockRepo) AtomicRecharge(_ context.Context, userID, orderID int64, amount float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.orders {
		if o.ID == orderID {
			if o.Status == "paid" {
				return nil
			}
			o.Status = "paid"
			u, ok := r.users[userID]
			if !ok {
				return errors.New("user not found")
			}
			u.Balance += amount
			return nil
		}
	}
	return errors.New("order not found")
}

func TestChargeGenerationIdempotent(t *testing.T) {
	repo := newMockRepo()
	_ = repo.CreateUser(context.Background(), 1, 10.00)
	svc := NewService(repo)

	ctx := context.Background()
	if err := svc.ChargeGeneration(ctx, 1, 100); err != nil {
		t.Fatalf("first charge failed: %v", err)
	}

	if err := svc.ChargeGeneration(ctx, 1, 100); err != ErrGenerationAlreadyCharged {
		t.Fatalf("expected ErrGenerationAlreadyCharged, got %v", err)
	}
}

func TestChargeGenerationUsesFreeQuota(t *testing.T) {
	repo := newMockRepo()
	_ = repo.CreateUser(context.Background(), 1, 10.00)
	repo.users[1].FreeQuota = 1

	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.ChargeGeneration(ctx, 1, 200); err != nil {
		t.Fatalf("charge failed: %v", err)
	}

	user, err := repo.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if user.FreeQuota != 0 {
		t.Fatalf("expected free_quota=0, got %d", user.FreeQuota)
	}
	if user.Balance != 10.00 {
		t.Fatalf("expected balance unchanged at 10.00, got %.2f", user.Balance)
	}

	trans, err := repo.GetTransactionByGeneration(ctx, 200)
	if err != nil {
		t.Fatalf("get transaction failed: %v", err)
	}
	if trans == nil {
		t.Fatal("expected transaction to exist")
	}
	if trans.Amount != 0 {
		t.Fatalf("expected transaction amount=0 for free quota, got %.2f", trans.Amount)
	}
}

func TestChargeGenerationDeductsBalance(t *testing.T) {
	repo := newMockRepo()
	_ = repo.CreateUser(context.Background(), 1, 10.00)
	repo.users[1].FreeQuota = 0

	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.ChargeGeneration(ctx, 1, 300); err != nil {
		t.Fatalf("charge failed: %v", err)
	}

	user, err := repo.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if user.Balance != 0.00 {
		t.Fatalf("expected balance=0.00, got %.2f", user.Balance)
	}

	trans, err := repo.GetTransactionByGeneration(ctx, 300)
	if err != nil {
		t.Fatalf("get transaction failed: %v", err)
	}
	if trans == nil {
		t.Fatal("expected transaction to exist")
	}
	if trans.Amount != 10.00 {
		t.Fatalf("expected transaction amount=10.00, got %.2f", trans.Amount)
	}
}

func TestHandlePaymentCallbackIdempotent(t *testing.T) {
	repo := newMockRepo()
	_ = repo.CreateUser(context.Background(), 1, 0)
	repo.orders = append(repo.orders, &Order{
		ID:        1,
		UserID:    1,
		OrderNo:   "ORD001",
		Amount:    50.00,
		Status:    "pending",
		WxOrderNo: "wx-123",
	})

	svc := NewService(repo)
	ctx := context.Background()

	if err := svc.HandlePaymentCallback(ctx, "wx-123"); err != nil {
		t.Fatalf("first callback failed: %v", err)
	}

	user, err := repo.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if user.Balance != 50.00 {
		t.Fatalf("expected balance=50.00 after first callback, got %.2f", user.Balance)
	}

	if err := svc.HandlePaymentCallback(ctx, "wx-123"); err != nil {
		t.Fatalf("second callback failed: %v", err)
	}

	user, err = repo.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if user.Balance != 50.00 {
		t.Fatalf("expected balance still 50.00 after second callback, got %.2f", user.Balance)
	}
}
