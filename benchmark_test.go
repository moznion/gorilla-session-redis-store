package redistore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
)

func BenchmarkMoznionRedistore(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	store := MakeDefaultWithGoRedisV9(client)
	store.KeyPrefix = "session:"
	store.NoWaitWritingMode = true

	defer func() {
		_ = store.Close(context.Background())
	}()
	req, _ := http.NewRequest("GET", "http://www.example.com", nil)
	w := httptest.NewRecorder()

	session := sessions.NewSession(store, "hello")
	for n := 0; n < b.N; n++ {
		_ = session.Save(req, w)
	}
}

func BenchmarkRbcervillaRedisstore(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	store, _ := redisstore.NewRedisStore(context.Background(), client)
	defer func() {
		_ = store.Close()
	}()
	req, _ := http.NewRequest("GET", "http://www.example.com", nil)
	w := httptest.NewRecorder()

	session := sessions.NewSession(store, "hello")
	for n := 0; n < b.N; n++ {
		_ = session.Save(req, w)
	}
}
