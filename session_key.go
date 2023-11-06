package redistore

type SessionKey struct {
	KeyPrefix string
	SessionID string
}

func (k *SessionKey) ToString() string {
	return k.KeyPrefix + k.SessionID
}
