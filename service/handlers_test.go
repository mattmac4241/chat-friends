package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var (
	formatter = render.New(render.Options{
		IndentJSON: true,
	})
	request  *http.Request
	recorder *httptest.ResponseRecorder
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
			return nil
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

func (t *testDatabase) getFriendsByUserID(userID uint) ([]FriendRequest, error) {
	var requests []FriendRequest
	for _, request := range t.requests {
		if request.UserFromID == userID || request.UserToID == userID {
			requests = append(requests, request)
		}
	}
	return requests, nil
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

func TestRejectRequestHandlerWithoutValidRequest(t *testing.T) {
	database := &testDatabase{}
	server := MakeTestServer(database)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("PUT", "/friends/1/reject", nil)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected %v; received %v", http.StatusNotFound, recorder.Code)
	}
}

func TestRejectRequestHandlerWithValidRequest(t *testing.T) {
	database := &testDatabase{}
	database.insertFriendRequest(FriendRequest{ID: 1, UserFromID: 1, UserToID: 2})

	server := MakeTestServer(database)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("PUT", "/friends/1/reject", nil)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected %v; received %v", http.StatusOK, recorder.Code)
	}

	if database.requests[0].RejectedAt.Unix() <= 0 {
		t.Error("Expected the reqeust to be rejected")
	}
}

func TestAcceptRequestHandlerWithoutValidRequest(t *testing.T) {
	database := &testDatabase{}
	server := MakeTestServer(database)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("PUT", "/friends/1/accept", nil)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("Expected %v; received %v", http.StatusNotFound, recorder.Code)
	}
}

func TestAcceptRequestHandlerWithValidRequest(t *testing.T) {
	database := &testDatabase{}
	database.insertFriendRequest(FriendRequest{ID: 1, UserFromID: 1, UserToID: 2})

	server := MakeTestServer(database)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("PUT", "/friends/1/accept", nil)
	server.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected %v; received %v", http.StatusOK, recorder.Code)
	}

	if database.requests[0].AcceptedAt.Unix() <= 0 {
		t.Error("Expected the reqeust to be rejected")
	}
}

func TestGetFriendsHandlerWithoutValidToken(t *testing.T) {
	database := &testDatabase{}

	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(getFriendsHandler(formatter, database)))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Errorf("Error in POST to registerUserHandler: %v", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Error("No auth token sent should have stopped it.")
	}
}

func TestGetFriendsHandlerWithoutAnyFriends(t *testing.T) {
	var friendRequests []FriendRequest

	database := &testDatabase{}
	database.redis = make(map[string]string)
	database.redisSetValue("TEST", "1", time.Since(time.Now()))

	database.insertFriendRequest(FriendRequest{
		UserFromID: 2,
		UserToID:   3,
		AcceptedAt: time.Now(),
	})

	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(getFriendsHandler(formatter, database)))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", "TEST")

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error in POST to registerUserHandler: %v", err)
	}

	defer resp.Body.Close()
	payload, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(payload, &friendRequests)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected %v; received %v", http.StatusOK, resp.StatusCode)
	}

	if len(friendRequests) > 0 {
		t.Errorf("Expected length to be 2 but instead got %d", len(friendRequests))
	}
}

func TestGetFriendsHandlerWithFriends(t *testing.T) {
	var friendRequests []FriendRequest

	database := &testDatabase{}
	database.redis = make(map[string]string)
	database.redisSetValue("TEST", "1", time.Since(time.Now()))

	database.insertFriendRequest(FriendRequest{
		UserFromID: 2,
		UserToID:   3,
		AcceptedAt: time.Now(),
	})

	database.insertFriendRequest(FriendRequest{
		UserFromID: 1,
		UserToID:   3,
		AcceptedAt: time.Now(),
	})

	database.insertFriendRequest(FriendRequest{
		UserFromID: 1,
		UserToID:   2,
		AcceptedAt: time.Now(),
	})

	client := &http.Client{}
	server := httptest.NewServer(http.HandlerFunc(getFriendsHandler(formatter, database)))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", "TEST")

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Error in POST to registerUserHandler: %v", err)
	}

	defer resp.Body.Close()
	payload, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(payload, &friendRequests)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected %v; received %v", http.StatusOK, resp.StatusCode)
	}

	if len(friendRequests) != 2 {
		t.Errorf("Expected length to be 2 but instead got %d", len(friendRequests))
	}
}

func MakeTestServer(database *testDatabase) *negroni.Negroni {
	server := negroni.New()
	mx := mux.NewRouter()
	initRoutes(mx, formatter, database)
	server.UseHandler(mx)
	return server
}
