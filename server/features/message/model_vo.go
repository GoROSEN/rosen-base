package message

import "time"

type MessageVO struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"time"`
	FromID       uint      `json:"fromId"`
	DestUserID   uint      `json:"destUserId,omitempty"`
	DestmemberID uint      `json:"destMemberId,omitempty"`
	Title        string    `json:"title"`
	Conent       string    `json:"content"`
	Unread       bool      `json:"unread"`
}
