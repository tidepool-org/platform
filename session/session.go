package session

type Session struct {
	ID        string `json:"-" bson:"_id"`
	IsServer  bool   `json:"isServer" bson:"isServer"`
	ServerID  string `json:"-" bson:"serverId,omitempty"`
	UserID    string `json:"userId,omitempty" bson:"userId,omitempty"`
	Duration  int64  `json:"-" bson:"duration"`
	ExpiresAt int64  `json:"-" bson:"expiresAt"`
	CreatedAt int64  `json:"-" bson:"createdAt"`
	Time      int64  `json:"-" bson:"time"`
}
