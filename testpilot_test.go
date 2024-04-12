package testpilot

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func Test_TestPlan(t *testing.T) {
	server := MockServer()
	defer server.Close()
	p := NewPlan(t, "User Test")
	p.Request("POST", server.URL+"/").
		Body(JSON(User{ID: 1, Name: "Max"})).
		Headers(map[string]string{
			"Content-Type": "application/json",
		}).
		Store("user").
		Expect().
		Status(201).
		Body(AssertPath(".id", func(val int) error {
			if val != 1 {
				return errors.New("expected 1 got " + strconv.Itoa(val))
			}
			return nil
		}))
	p.Request("GET", server.URL).
		Expect().
		Status(200).
		Body(AssertPath(".0.id", func(val int) error {
			if val != 1 {
				return errors.New("expected 1 got " + strconv.Itoa(val))
			}
			return nil
		}))
	p.Request("GET", server.URL+"/{user.id}").Expect().Status(200).Body(ResponseComparer(p))
	p.Request("GET", server.URL+"/2").Expect().Status(404)
	p.Request("GET", server.URL+"/ping").Expect().Status(200).Body(AssertEqual("pong"))
	p.Run()

}

func ResponseComparer(p *TestPlan) func(body []byte) error {
	return func(body []byte) error {
		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return err
		}
		resp := p.ResponseForKey("user")
		var storedUser User
		if err := json.Unmarshal(resp, &storedUser); err != nil {
			return err
		}
		if user.ID != storedUser.ID {
			return errors.New("expected " + strconv.Itoa(storedUser.ID) + " got " + strconv.Itoa(user.ID))
		}
		return nil
	}
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type MockStore struct {
	Store []User
}

var mockStore = MockStore{
	Store: []User{},
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(mockStore.Store)
	w.Write(data)
}
func fetchUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	for _, user := range mockStore.Store {
		if strconv.Itoa(user.ID) == id {
			data, _ := json.Marshal(user)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error": "user not found"}`))
}
func addUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	mockStore.Store = append(mockStore.Store, user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	data, _ := json.Marshal(user)
	w.Write(data)
}
func pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))

}
func MockServer() *httptest.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("GET /", listUsers)
	mux.HandleFunc("POST /", addUser)
	mux.HandleFunc("GET /{id}", fetchUser)
	mux.HandleFunc("GET /ping", pong)
	return httptest.NewServer(mux)
}
