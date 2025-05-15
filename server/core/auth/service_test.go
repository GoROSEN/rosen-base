package auth

import (
	"testing"

	"github.com/GoROSEN/rosen-apiserver/core/user"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ResourceStub struct {
	db          *gorm.DB
	redisServer *miniredis.Miniredis
	redisClient *redis.Client
}

func (r *ResourceStub) Close() {

	// r.db.Close()
	r.redisClient.Close()
	r.redisServer.Close()
}

func createService() (*Service, *ResourceStub) {

	db, _ := gorm.Open(sqlite.Open("/tmp/test.db"), &gorm.Config{})
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	redis := redis.NewClient(&redis.Options{
		Addr:     s.Addr(), // use default Addr
		Password: "",       // no password set
		DB:       0,        // use default DB
	})

	as := NewAuthService(redis, db)
	user.MigrateDB(db)

	return as, &ResourceStub{db, s, redis}
}

func TestAuth(t *testing.T) {

	as, stub := createService()
	defer stub.Close()

	//create token
	token, err := as.AuthUser("Olivia", "JGNkfMqf")
	if err != nil {
		t.Error(err)
	} else if token == nil {
		t.Error("token is nil")
	}

	// verify token
	if _, _, _, err = as.VerifyUserToken(token.Token); err != nil {
		t.Errorf("verify token failed: %v", err)
	}
}

func TestRemoveToken(t *testing.T) {

	as, stub := createService()
	defer stub.Close()

	//create token
	token, err := as.AuthUser("Olivia", "JGNkfMqf")
	if err != nil {
		t.Error(err)
	} else if token == nil {
		t.Error("token is nil")
	}
	as.RemoveToken(token.Token)
	// verify token
	if _, _, _, err = as.VerifyUserToken(token.Token); err == nil {
		t.Errorf("expect verifying failed, got a success")
	}
}
