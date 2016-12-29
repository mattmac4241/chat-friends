package service

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq" // needed
)

var DB *sql.DB

type Database interface {
	getFriendRequestByUserFromAndTo(userFrom, userTo uint) (FriendRequest, error)
	insertFriendRequest(request FriendRequest) error
	updateFriendRequest(request FriendRequest) error
	redisGetValue(key string) (string, error)
	redisSetValue(key, value string, seconds time.Duration) error
	getFriendRequestByID(requestID uint) (FriendRequest, error)
}

type dataHandler struct{}

func (d *dataHandler) getFriendRequestByUserFromAndTo(userFrom, userTo uint) (FriendRequest, error) {
	var request FriendRequest
	err := DB.QueryRow(`SELECT ID, USER_FROM_ID, USER_TO_ID, CREATED_AT, ACCEPTED_AT,
		REJECTED_ID FROM friend_requests WHERE (user_id_from=$1 OR user_id_to=$1)
		AND (user_id_from=$2 or user_id_to=$2);`, userFrom, userTo).Scan(request.ID,
		request.UserFromID, request.UserToID, request.CreatedAt, request.AcceptedAt,
		request.RejectedAt)
	return request, err
}

func (d *dataHandler) getFriendRequestByID(requestID uint) (FriendRequest, error) {
	var request FriendRequest
	err := DB.QueryRow(`SELECT ID, USER_FROM_ID, USER_TO_ID, CREATED_AT, ACCEPTED_AT,
		REJECTED_ID FROM friend_requests WHERE id=$1;`, requestID).Scan(request.ID,
		request.UserFromID, request.UserToID, request.CreatedAt, request.AcceptedAt,
		request.RejectedAt)
	return request, err
}

func (d *dataHandler) insertFriendRequest(request FriendRequest) error {
	var lastInsertID uint
	err := DB.QueryRow(`INSERT INTO friend_requests (USER_FROM_ID, USER_TO_ID)
			VALUES($1, $2) returning id;`, request.UserFromID, request.UserToID).Scan(lastInsertID)
	return err
}

func (d *dataHandler) updateFriendRequest(request FriendRequest) error {
	var lastInsertID uint
	err := DB.QueryRow(`UPDATE friend_requests SET accpected_at=$1, rejected_at=$2
		WHERE ID=$3 returning id;`, request.AcceptedAt, request.RejectedAt,
		request.ID).Scan(lastInsertID)
	return err
}

func (d *dataHandler) redisGetValue(key string) (string, error) {
	return REDIS.Get(key).Result()
}

func (d *dataHandler) redisSetValue(key, value string, seconds time.Duration) error {
	return REDIS.Set(key, value, seconds).Err()
}

//InitDatabase setup db connection
func InitDatabase(dbinfo string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbinfo+" sslmode=disable")
	return db, err
}
