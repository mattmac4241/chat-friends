package service

import (
	"testing"
	"time"
)

func TestFriendRequestAccept(t *testing.T) {
	friendRequest := FriendRequest{
		UserFromID: 1,
		UserToID:   2,
	}
	friendRequest.accept()
	if friendRequest.AcceptedAt.Equal(time.Unix(0, 0)) {
		t.Error("Accept failed")
	}
}

func TestFriendRequestReject(t *testing.T) {
	friendRequest := FriendRequest{
		UserFromID: 1,
		UserToID:   2,
	}
	friendRequest.reject()
	if friendRequest.RejectedAt.Equal(time.Unix(0, 0)) {
		t.Error("Reject failed")
	}
}

func TestFriendRequestSave(t *testing.T) {
	friendRequest := FriendRequest{
		UserFromID: 1,
		UserToID:   2,
	}
	friendRequest.save()
	if friendRequest.CreatedAt.Equal(time.Unix(0, 0)) {
		t.Error("Save failed")
	}
}

func TestAddFriend(t *testing.T) {
	request := AddFriend(1, 2)
	if request.UserFromID != 1 || request.UserToID != 2 {
		t.Error("Friend request creation failed")
	}
}
