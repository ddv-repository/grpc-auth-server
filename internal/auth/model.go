package auth

import (
	"time"
)

type AccessToken struct {
	Token     string        `bson:"token"`
	ClientID  string        `bson:"client_id"`
	UserID    string        `bson:"user_id"`
	Scope     []string      `bson:"scope"`
	IPAddress []string      `bson:"ip_address"`
	ExpiredIn time.Duration `bson:"expired_in"`
	OneTime   bool          `bson:"one_time"`
	TokenType string        `bson:"token_type"`
}
