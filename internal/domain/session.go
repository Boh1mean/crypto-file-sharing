package domain

import "time"

type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (s Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

type Challenge struct {
	Nonce     []byte
	UserID    int
	ExpiresAt time.Time
}

func (c Challenge) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}
