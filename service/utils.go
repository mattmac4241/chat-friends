package service

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
)

func getUserFromHeader(req *http.Request, data Database) (uint, error) {
	key := req.Header.Get("Authorization")
	user, err := data.redisGetValue(key)
	if err != nil {
		return uint(0), errors.New("Failed to find user")
	}
	userID, _ := strconv.ParseUint(user, 10, 32)
	return uint(userID), nil
}

func getFriendRequest(userIDFrom, userIDTo uint, database Database) (FriendRequest, error) {
	request, err := database.getFriendRequestByUserFromAndTo(userIDFrom, userIDTo)
	if err == nil {
		return FriendRequest{}, errors.New("No friend request found")
	}
	return request, nil
}

func hasFriendRequest(userIDFrom, userIDTo uint, database Database) bool {
	request, err := getFriendRequest(userIDFrom, userIDTo, database)
	return request != FriendRequest{} && err == nil
}

func conevertRowsToRequests(rows *sql.Rows) []FriendRequest {
	var requests []FriendRequest
	defer rows.Close()
	for rows.Next() {
		var request FriendRequest
		err := rows.Scan(request.ID,
			request.UserFromID, request.UserToID, request.CreatedAt,
			request.AcceptedAt, request.RejectedAt)
		if err == nil {
			requests = append(requests, request)
		}
	}
	return requests
}
