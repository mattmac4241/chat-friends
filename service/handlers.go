package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

func postAddFriend(formatter *render.Render, database Database) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, err := getUserFromHeader(req, database)
		if err != nil {
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}
		var request FriendRequest
		payload, _ := ioutil.ReadAll(req.Body)
		err = json.Unmarshal(payload, &request)
		if err != nil || (request == FriendRequest{}) {
			formatter.Text(w, http.StatusBadRequest, "Failed to parse request.")
			return
		}
		if hasFriendRequest(userID, request.UserToID, database) == true {
			formatter.Text(w, http.StatusBadRequest, "Request already exists.")
			return
		}
		request = AddFriend(userID, request.UserToID)
		err = database.insertFriendRequest(request)
		if err != nil {
			formatter.Text(w, http.StatusBadRequest, "Failed to add request.")
			return
		}
		formatter.Text(w, http.StatusCreated, "Request succesfully created.")
	}
}

func rejectRequest(formatter *render.Render, database Database) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["request_id"]
		if key == "" {
			formatter.JSON(w, http.StatusNotFound, "No request id sent.")
			return
		}
		requestID, _ := strconv.ParseUint(key, 10, 32)
		request, err := database.getFriendRequestByID(uint(requestID))
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, "No request found.")
			return
		}
		request.reject()
		err = database.updateFriendRequest(request)
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, "Failed to update request.")
			return
		}
		formatter.JSON(w, http.StatusOK, "Request rejected")
	}
}

func acceptRequest(formatter *render.Render, database Database) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		key := vars["request_id"]
		if key == "" {
			formatter.JSON(w, http.StatusNotFound, "No request id sent.")
			return
		}
		requestID, _ := strconv.ParseUint(key, 10, 32)
		request, err := database.getFriendRequestByID(uint(requestID))
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, "No request found.")
			return
		}
		request.accept()
		err = database.updateFriendRequest(request)
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, "Failed to update request.")
			return
		}
		formatter.JSON(w, http.StatusOK, "Request rejected")
	}
}

func getFriendsHandler(formatter *render.Render, database Database) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		userID, err := getUserFromHeader(req, database)
		if err != nil {
			formatter.JSON(w, http.StatusBadRequest, err)
			return
		}
		requests, err := database.getFriendRequestByID(userID)
		if err != nil {
			formatter.JSON(w, http.StatusNotFound, "Failed to get friends.")
			return
		}
		formatter.JSON(w, http.StatusOK, requests)
	}
}
