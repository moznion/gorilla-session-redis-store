package redistore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
)

const (
	redisAddr = "localhost:6379"
)

func TestNew(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	store := MakeDefaultWithGoRedisV9(client)
	store.KeyPrefix = "prefix:"

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}
	if session.IsNew == false {
		t.Fatal("session is not new")
	}
}

func TestSave(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	store := MakeDefaultWithGoRedisV9(client)
	store.KeyPrefix = "prefix:"

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	session.Values["key"] = "value"
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save: ", err)
	}
}

func TestDelete(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	store := MakeDefaultWithGoRedisV9(client)
	store.KeyPrefix = "prefix:"

	req, err := http.NewRequest("GET", "http://www.example.com", nil)
	if err != nil {
		t.Fatal("failed to create request", err)
	}
	w := httptest.NewRecorder()

	session, err := store.New(req, "hello")
	if err != nil {
		t.Fatal("failed to create session", err)
	}

	session.Values["key"] = "value"
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to save session: ", err)
	}

	session.Options.MaxAge = -1
	err = session.Save(req, w)
	if err != nil {
		t.Fatal("failed to delete session: ", err)
	}
}

func TestClose(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	cmd := client.Ping(context.Background())
	err := cmd.Err()
	if err != nil {
		t.Fatal("connection is not opened")
	}

	store := MakeDefaultWithGoRedisV9(client)
	store.KeyPrefix = "prefix:"

	err = store.Close(context.Background())
	if err != nil {
		t.Fatal("failed to close")
	}

	cmd = client.Ping(context.Background())
	if cmd.Err() == nil {
		t.Fatal("connection is properly closed")
	}
}
