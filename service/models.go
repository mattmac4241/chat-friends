package service

import (
	"time"
)

//FriendRequest describes a friend request
type FriendRequest struct {
	ID         uint      `json:"id"`
	UserFromID uint      `json:"user_from_id"`
	UserToID   uint      `json:"user_to_id"`
	CreatedAt  time.Time `json:"created_at"`
	AcceptedAt time.Time `json:"accepted_at"`
	RejectedAt time.Time `json:"rejected_at"`
}

func (f *FriendRequest) accept() {
	f.AcceptedAt = time.Now()
}

func (f *FriendRequest) reject() {
	f.RejectedAt = time.Now()
}

func (f *FriendRequest) save() {
	f.CreatedAt = time.Now()
}

//AddFriend creates a FriendRequest
func AddFriend(friendFrom, friendTo uint) FriendRequest {
	request := FriendRequest{
		UserFromID: friendFrom,
		UserToID:   friendTo,
	}
	request.save()
	return request
}
