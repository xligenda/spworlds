package spwmini

type User struct {
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	MinecraftUUID string   `json:"minecraftUUID"`
	Timestamp     int64    `json:"timestamp"`
	Roles         []string `json:"roles"`
	IsAdmin       bool     `json:"isAdmin"`
	Hash          string   `json:"hash"`
}
