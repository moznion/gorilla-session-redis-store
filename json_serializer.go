package redistore

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/gorilla/sessions"
)

type JSONSerializer struct {
	JSONSerde JSONSerde
}

type JSONSerde interface {
	Marshal(ctx context.Context, session *sessions.Session) ([]byte, error)
	Unmarshal(ctx context.Context, serialized []byte, session *sessions.Session) error
}

func makeDefaultJSONSerializer() *JSONSerializer {
	return &JSONSerializer{JSONSerde: &defaultJSONSerde{}}
}

func (s JSONSerializer) Serialize(ctx context.Context, session *sessions.Session) ([]byte, error) {
	serialized, err := s.JSONSerde.Marshal(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize the session by JSONSerializer: %w", err)
	}
	return serialized, nil
}

func (s JSONSerializer) Deserialize(ctx context.Context, serialized []byte, session *sessions.Session) error {
	err := s.JSONSerde.Unmarshal(ctx, serialized, session)
	if err != nil {
		return fmt.Errorf("failed to deserialize the session by JSONSerializer: %w", err)
	}
	return nil
}

type defaultJSONSerde struct {
}

func (djs *defaultJSONSerde) Marshal(ctx context.Context, session *sessions.Session) ([]byte, error) {
	v := session.Values
	m := make(map[string]interface{}, len(v))
	for k, v := range v {
		ks, ok := k.(string)
		if !ok {
			return nil, fmt.Errorf("cannot serialize the session into JSON due to the given session values has non-string key value; %v", k)
		}
		m[ks] = v
	}

	serialized, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize the session into JSON by defaultJSONSerde: %w", err)
	}
	return serialized, nil
}

func (djs *defaultJSONSerde) Unmarshal(ctx context.Context, serialized []byte, session *sessions.Session) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(serialized, &m)
	if err != nil {
		return fmt.Errorf("failed to deserialize the session from JSON by defaultJSONSerde: %w", err)
	}

	for k, v := range m {
		session.Values[k] = v
	}
	return nil
}
