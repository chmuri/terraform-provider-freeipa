package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_LoginAndPassword(t *testing.T) {
	// 1. Create TLS Mock Server (since FreeIPA calls require HTTPS)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ipa/session/login_password":
			if r.Method != "POST" {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			err := r.ParseForm()
			if err != nil {
				t.Errorf("failed to parse form: %v", err)
			}
			if r.FormValue("user") == "admin" && r.FormValue("password") == "SecretPassword" {
				// Success, set cookie and return OK
				http.SetCookie(w, &http.Cookie{Name: "ipa_session", Value: "session-cookie-123"})
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}

		case "/ipa/session/json":
			if r.Method != "POST" {
				t.Errorf("expected POST method, got %s", r.Method)
			}
			// Read JSON-RPC request
			var req RPCRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				t.Errorf("failed to decode request: %v", err)
			}

			if req.Method == "user_show" {
				responseJSON := `{
					"result": {
						"result": {
							"uid": ["jdoe"],
							"givenname": ["John"],
							"sn": ["Doe"]
						}
					},
					"error": null,
					"id": 1
				}`
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(responseJSON))
			} else {
				w.WriteHeader(http.StatusNotFound)
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Extract the host and port from the mock server URL
	mockHost := strings.TrimPrefix(server.URL, "https://")

	cfg := &Config{
		Host:       mockHost,
		Insecure:   true, // Required to bypass TLS verification of the test server self-signed certificate
		AuthMethod: AuthPassword,
		Username:   "admin",
		Password:   "SecretPassword",
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test Login
	err = c.Login()
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Test Call (user_show)
	var result struct {
		Result struct {
			Uid       []string `json:"uid"`
			Givenname []string `json:"givenname"`
			Sn        []string `json:"sn"`
		} `json:"result"`
	}

	err = c.Call(context.Background(), "user_show", []string{"jdoe"}, nil, &result)
	if err != nil {
		t.Fatalf("API call failed: %v", err)
	}

	if len(result.Result.Uid) == 0 || result.Result.Uid[0] != "jdoe" {
		t.Errorf("expected uid 'jdoe', got %v", result.Result.Uid)
	}
	if len(result.Result.Givenname) == 0 || result.Result.Givenname[0] != "John" {
		t.Errorf("expected givenname 'John', got %v", result.Result.Givenname)
	}
}

func TestClient_RPCError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ipa/session/json" {
			responseJSON := `{
				"result": null,
				"error": {
					"code": 4002,
					"message": "some_name: search delete etc. not found",
					"name": "NotFoundError"
				},
				"id": 1
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	mockHost := strings.TrimPrefix(server.URL, "https://")

	cfg := &Config{
		Host:       mockHost,
		Insecure:   true,
		AuthMethod: AuthPassword,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Make call that should fail with 4002 error
	err = c.Call(context.Background(), "user_show", []string{"nonexistent"}, nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	rpcErr, ok := err.(*RPCError)
	if !ok {
		t.Fatalf("expected error of type *RPCError, got %T: %v", err, err)
	}

	if rpcErr.Code != 4002 {
		t.Errorf("expected error code 4002, got %d", rpcErr.Code)
	}
	if rpcErr.Name != "NotFoundError" {
		t.Errorf("expected error name NotFoundError, got %s", rpcErr.Name)
	}
}

