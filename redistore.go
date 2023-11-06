package redistore

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
)

type Redistore struct {
	RedisClient       RedisClient
	Options           sessions.Options
	KeyGenerator      KeyGenerator
	KeyPrefix         string
	Serializer        Serializer
	NoWaitWritingMode bool
}

func MakeDefaultWithGoRedisV9(goRedisV9Client *redis.Client) *Redistore {
	return &Redistore{
		RedisClient: &GoRedisV9Client{Client: goRedisV9Client},
		Options: sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		KeyGenerator:      &DefaultRandomKeyGenerator{},
		KeyPrefix:         "",
		Serializer:        makeDefaultJSONSerializer(),
		NoWaitWritingMode: false,
	}
}

// Get returns a session for the given name after adding it to the registry.
func (s *Redistore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
func (s *Redistore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := s.Options
	session.Options = &opts
	session.IsNew = true

	c, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	session.ID = c.Value

	valueExists, err := s.loadSession(r.Context(), session)
	if err != nil {
		return nil, err
	}
	session.IsNew = !valueExists

	return session, err
}

func (s *Redistore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		key := &SessionKey{
			KeyPrefix: s.KeyPrefix,
			SessionID: session.ID,
		}
		if s.NoWaitWritingMode {
			go s.RedisClient.Del(r.Context(), key)
		} else {
			err := s.RedisClient.Del(r.Context(), key)
			if err != nil {
				return fmt.Errorf("failed to delete a session from redis; key = %s: %w", key.ToString(), err)
			}
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	allowOverwrite := true
	if session.ID == "" {
		id, err := s.KeyGenerator.GenerateKey()
		if err != nil {
			return errors.New("failed to generate a session id")
		}
		session.ID = id
		allowOverwrite = false
	}

	if s.NoWaitWritingMode {
		go s.storeSession(r.Context(), session, allowOverwrite)
	} else {
		err := s.storeSession(r.Context(), session, allowOverwrite)
		if err != nil {
			return fmt.Errorf("failed to store the session to redis: %w", err)
		}
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), session.ID, session.Options))
	return nil
}

// Close closes the Redis store
func (s *Redistore) Close(ctx context.Context) error {
	return s.RedisClient.Close(ctx)
}

func (s *Redistore) storeSession(ctx context.Context, session *sessions.Session, allowOverwrite bool) error {
	b, err := s.Serializer.Serialize(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to serialize the session on session storing: %w", err)
	}

	if allowOverwrite {
		err := s.RedisClient.Set(ctx, &SessionKey{
			KeyPrefix: s.KeyPrefix,
			SessionID: session.ID,
		}, b, time.Duration(session.Options.MaxAge)*time.Second)
		if err != nil {
			return fmt.Errorf("failed to set the session to redis: %w", err)
		}
		return nil
	}

	succeeded, err := s.RedisClient.SetNX(ctx, &SessionKey{
		KeyPrefix: s.KeyPrefix,
		SessionID: session.ID,
	}, b, time.Duration(session.Options.MaxAge)*time.Second)
	if err != nil {
		return fmt.Errorf("failed to setNX the session to redis: %w", err)
	}
	if !succeeded {
		return fmt.Errorf("failed to setNX the session because of the duplicated ID: %s", session.ID)
	}
	return nil
}

func (s *Redistore) loadSession(ctx context.Context, session *sessions.Session) (bool, error) {
	b, exists, err := s.RedisClient.Get(ctx, &SessionKey{
		KeyPrefix: s.KeyPrefix,
		SessionID: session.ID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get a session value from redis: %w", err)
	}
	if !exists {
		return false, nil
	}

	err = s.Serializer.Deserialize(ctx, b, session)
	if err != nil {
		return false, fmt.Errorf("failed to deserialize the fetched session: %w", err)
	}
	return true, nil
}

type Serializer interface {
	Serialize(context.Context, *sessions.Session) ([]byte, error)
	Deserialize(context.Context, []byte, *sessions.Session) error
}
