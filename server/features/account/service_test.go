package account

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newService() *AccountService {

	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	MigrateDB(db)
	s := NewAccountService(db)
	return s
}

func TestNewAccount(t *testing.T) {

	s := newService()
	// defer s.db.Close()
	a := s.NewAccount(1, "RMB")
	if a == nil {
		t.Error("Cannot create account")
	}
}

func TestTransfer(t *testing.T) {

	s := newService()
	// defer s.db.Close()
	a := s.NewAccount(1, "RMB")
	if a == nil {
		t.Error("Cannot create account a")
	}
	_, err := s.IncreaseAvailable(a, 100)
	if err != nil {
		t.Error("IncreaseAvailable error: ", err)
	}

	b := s.GetAccountByID(a.ID)
	if b == nil {
		t.Errorf("Cannot find the account %v", a.ID)
	}
	if b.Available != 100 {
		t.Errorf("Expect account available 100, got %v", b.Available)
	}

	c := s.NewAccount(2, "RMB")
	if c == nil {
		t.Error("Cannot create account c")
	}
	_, err = s.Transfer(a, c, 50)
	if err != nil {
		t.Error("Transfer error: ", err)
	}
	if a.Available != 50 || c.Available != 50 {
		t.Errorf("Expect a.available = 50, got %v", a.Available)
		t.Errorf("Expect c.available = 50, got %v", c.Available)
	}
}

func TestLock(t *testing.T) {

	s := newService()
	// defer s.db.Close()
	a := s.NewAccount(1, "RMB")
	if a == nil {
		t.Error("Cannot create account a")
	}
	_, err := s.IncreaseLocked(a, 1000)
	if err != nil {
		t.Error("IncreaseLocked error: ", err)
	}
	s.Unlock(a, 100)
	if a.Available != 100 || a.Locked != 900 {
		t.Errorf("Expect a.available = 100, got %v", a.Available)
		t.Errorf("Expect a.locked = 900, got %v", a.Locked)
	}
	s.Lock(a, 10)
	if a.Available != 90 || a.Locked != 910 {
		t.Errorf("Expect a.available = 90, got %v", a.Available)
		t.Errorf("Expect a.locked = 910, got %v", a.Locked)
	}
	s.DecreaseLocked(a, 90)
	if a.Locked != 820 {
		t.Errorf("Expect a.locked = 820, got %v", a.Locked)
	}
}

func TestFreeze(t *testing.T) {

	s := newService()
	// defer s.db.Close()
	a := s.NewAccount(1, "RMB")
	b := s.NewAccount(2, "RMB")
	if a == nil {
		t.Error("Cannot create account a")
	}
	_, err := s.IncreaseAvailable(a, 1000)
	if err != nil {
		t.Error("IncreaseAvailable error: ", err)
	}

	s.Freeze(a, 100)
	if a.Available != 900 || a.Frozen != 100 {
		t.Errorf("Expect a.available = 900, got %v", a.Available)
		t.Errorf("Expect a.frozen = 100, got %v", a.Frozen)
	}
	s.Unfreeze(a, b, 100)
	if a.Available != 900 || b.Available != 100 {
		t.Errorf("Expect a.available = 900, got %v", a.Available)
		t.Errorf("Expect b.available = 100, got %v", b.Available)
	}
}

func TestAccount(t *testing.T) {

	s := newService()
	// defer s.db.Close()
	a := s.Account(1, "RMB")
	if a == nil {
		t.Error("Cannot get account a")
	}
}
