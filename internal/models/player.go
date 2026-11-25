package models

import "github.com/gorilla/websocket"

type Player struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	RoomID string `json:"room_id"`

	Conn *websocket.Conn `json:"-"`

	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	PositionZ float64 `json:"position_z"`

	RotationX float64 `json:"rot_x"`
	RotationY float64 `json:"rot_y"`

	HP      int  `json:"hp"`
	IsAlive bool `json:"is_alive"`

	Weapon  string `json:"weapon"`
	Ammo    int    `json:"ammo"`
	MaxAmmo int    `json:"max_ammo"`

	IsShooting bool   `json:"is_shooting"`
	IsRunning  bool   `json:"is_running"`
	IsWalking  bool   `json:"is_walking"`
	IsWalkingr bool   `json:"is_walkingr"`
	IsWalkingl bool   `json:"is_walkingl"`
	IsJump     bool   `json:"is_jump"`
	Type       string `json:"type"`
	IsMatched  bool   `json:"-"`

	SpawnIndex int              `json:"spawn_index"`
	Send       chan interface{} `json:"-"`
}
type BaseMessage struct {
	Type string `json:"type"`
}

type HPUpdateMessage struct {
	Type     string `json:"type"`
	PlayerID string `json:"id"`
	HP       int    `json:"hp"`
}
type GameResultMessage struct {
	Type     string `json:"type"`
	WinnerID string `json:"winner_id"`
	LoserID  string `json:"loser_id"`
}
type ShootMessage struct {
	Type     string `json:"type"`
	PlayerID string `json:"player_id"`

	Hit bool `json:"hit"`

	TargetID string `json:"target_id"`

	Damage int `json:"damage"`
}
