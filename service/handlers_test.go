package service

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/unrolled/render"
)

var (
	formatter = render.New(render.Options{
		IndentJSON: true,
	})
)

type testDatabase struct {
	requests []FriendRequest
	redis    map[string]string
}

func (t *testDatabase) getFriendRequestByUserFromAndTo(userFrom, userTo uint) (FriendRequest, error) {
	for _, request := range t.requests {
		if (request.UserFromID == userFrom || request.UserToID == userFrom) &&
			(request.UserFromID == userTo || request.UserToID == userTo) {
			return request, nil
		}
	}

	return FriendRequest{}, errors.New("No request found")
}

func (t *testDatabase) insertFriendRequest(request FriendRequest) error {
	t.requests = append(t.requests, request)
	return nil
}

func (t *testDatabase) updateFriendRequest(request FriendRequest) error {
	for indx, searchedRequest := range t.requests {
		if (searchedRequest.UserFromID == request.UserFromID ||
			searchedRequest.UserToID == request.UserFromID) &&
			(searchedRequest.UserFromID == request.UserToID ||
				searchedRequest.UserToID == request.UserToID) {
			t.requests[indx] = request
		}
	}
	return errors.New("Request not found to update")
}
func (t *testDatabase) redisGetValue(key string) (string, error) {
	return t.redis[key], nil
}
func (t *testDatabase) redisSetValue(key, value string, seconds time.Duration) error {
	t.redis[key] = value
	return nil
}
func (t *testDatabase) getFriendRequestByID(requestID uint) (FriendRequest, error) {
	for _, request := range t.requests {
		if request.ID == requestID {
			return request, nil
		}
	}
	return FriendRequest{}, errors.New("Request not found")
}

func TestPostAddFriendHandlerWithoutAuthKey(t *testing.T) {
	database := &testDatabase{}

	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(postAddFriendHandler(formatter, database)))
	defer server.Close()

	body := []byte("this is not valid json")
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))

	if err != nil {
		t.Errorf("Error in creating POST request for registerUserHandler: %v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error in POST to registerUserHandler: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Error("No auth token sent should have stopped it.")
	}
}

func TestPostAddFriendHandlerInvalidJSON(t *testing.T) {
	database := &testDatabase{}
	database.redis = make(map[string]string)
	database.redisSetValue("TEST", "1", time.Since(time.Now()))

	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(postAddFriendHandler(formatter, database)))
	defer server.Close()

	body := []byte("this is not valid json")

	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "TEST")

	if err != nil {
		t.Errorf("Error in creating POST request for registerUserHandler: %v", err)
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		t.Errorf("Error in POST to registerUserHandler: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Sending invalid JSON should result in a bad request from server.")
	}
}

func TestPostAddFriendHandlerHandlerNotFriendRequest(t *testing.T) {
	database := &testDatabase{}
	database.redis = make(map[string]string)
	database.redisSetValue("TEST", "1", time.Since(time.Now()))

	client := &http.Client{}

	server := httptest.NewServer(http.HandlerFunc(postAddFriendHandler(formatter, database)))
	defer server.Close()

	body := []byte("{\"test\":\"Not comment.\"}")
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))

	if err != nil {
		t.Errorf("Error in creating second POST request for invalid data on create user: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "TEST")

	res, _ := client.Do(req)
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Error("Sending valid JSON but with incorrect or missing fields should result in a bad request and didn't.")
	}
}

func TestPostAddFriendHandlerHandlerSuccess(t *testing.T) {
	database := &testDatabase{}
	database.redis = make(map[string]string)
	database.redisSetValue("TEST", "1", time.Since(time.Now()))

	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(postAddFriendHandler(formatter, database)))
	defer server.Close()

	body := []byte("{\"user_to_id\": 2}")
	req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))

	if err != nil {
		t.Errorf("Error in creating second POST request for invalid data on create user: %v", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "TEST")

	resp, _ := client.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Error("Sending valid JSON but with incorrect or missing fields should result in a bad request and didn't.")
		return
	}
	if len(database.requests) <= 0 {
		t.Error("Request not added to database")
		return
	}
}
