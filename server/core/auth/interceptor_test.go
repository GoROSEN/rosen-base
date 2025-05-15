package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GoROSEN/rosen-apiserver/core/user"
	"github.com/GoROSEN/rosen-apiserver/core/utils"

	"github.com/alicebob/miniredis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"gorm.io/gorm"
)

func createInterceptor(r *gin.Engine) (*Interceptor, *ResourceStub) {

	db, _ := gorm.Open("sqlite3", "/tmp/test.db")
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	redis := redis.NewClient(&redis.Options{
		Addr:     s.Addr(), // use default Addr
		Password: "",       // no password set
		DB:       0,        // use default DB
	})

	NewController(r, redis, db)
	user.MigrateDB(db)

	i := NewInterceptor(redis, db)

	r.GET("/api/test/verifyToken", i.AuthInterceptor, func(ctx *gin.Context) {

		ctx.JSON(http.StatusOK, gin.H{"userID": ctx.GetInt("userID")})
	})

	return i, &ResourceStub{db, s, redis}
}

func testInterceptedRequest(req *http.Request, t *testing.T, expectCode int) map[string]interface{} {

	r := gin.Default()
	_, stub := createInterceptor(r)
	defer stub.Close()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if rr.Code != expectCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, expectCode)
		t.Errorf("error message: %v", rr.Body)
	}

	if expectCode != http.StatusOK {
		return nil
	}

	m, err := utils.CheckJSONResult(rr.Body.Bytes())

	if err != nil {
		t.Error(err)
	}
	return m
}

func TestInterceptorBadAuth(t *testing.T) {

	req, err := http.NewRequest("GET", "/api/test/verifyToken", nil)
	if err != nil {
		t.Fatal(err)
	}

	testInterceptedRequest(req, t, http.StatusBadRequest)
}

func TestInterceptorInvalidAuth(t *testing.T) {

	req, err := http.NewRequest("GET", "/api/test/verifyToken?token=INVALIDTOKEN", nil)
	if err != nil {
		t.Fatal(err)
	}

	testInterceptedRequest(req, t, http.StatusForbidden)
}
