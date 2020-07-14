package session

import "time"

type SessionRepository struct {
	ID        string    `json:"-" bson:"_id"`
	IsServer  bool      `json:"isServer" bson:"isServer"`
	ServerID  string    `json:"-" bson:"serverId,omitempty"`
	UserID    string    `json:"userId,omitempty" bson:"userId,omitempty"`
	Duration  int64     `json:"-" bson:"duration"`
	ExpiresAt time.Time `json:"-" bson:"expiresAt"`
	CreatedAt time.Time `json:"-" bson:"createdAt"`
	Time      time.Time `json:"-" bson:"time"`
}
