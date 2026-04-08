package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken(secret, "user-1", "admin")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID = %s, want user-1", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %s, want admin", claims.Role)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	_, err := ParseToken("secret", "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseTokenWrongSecret(t *testing.T) {
	token, _ := GenerateToken("secret1", "user-1", "user")
	_, err := ParseToken("secret2", token)
	if err == nil {
		t.Error("expected error for wrong secret")
	}
}

func TestFixedUUIDs(t *testing.T) {
	if AdminID == "" {
		t.Error("AdminID should not be empty")
	}
	if UserID == "" {
		t.Error("UserID should not be empty")
	}
	if AdminID == UserID {
		t.Error("AdminID and UserID should be different")
	}
}

func TestMiddlewareMissingToken(t *testing.T) {
	mw := Middleware("secret")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mw(handler).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareInvalidToken(t *testing.T) {
	mw := Middleware("secret")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	mw(handler).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareValidToken(t *testing.T) {
	mw := Middleware("secret")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := FromContext(r.Context())
		if claims == nil {
			t.Error("expected claims in context")
			return
		}
		if claims.UserID != "user-1" {
			t.Errorf("UserID = %s, want user-1", claims.UserID)
		}
		if claims.Role != "user" {
			t.Errorf("Role = %s, want user", claims.Role)
		}
		w.WriteHeader(http.StatusOK)
	})

	token, _ := GenerateToken("secret", "user-1", "user")
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestFromContextEmpty(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	claims := FromContext(r.Context())
	if claims != nil {
		t.Error("expected nil claims for empty context")
	}
}

func TestWithClaims(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	ctx := WithClaims(r.Context(), &Claims{UserID: "test", Role: "admin"})
	claims := FromContext(ctx)
	if claims == nil {
		t.Fatal("expected claims in context")
	}
	if claims.UserID != "test" {
		t.Errorf("UserID = %s, want test", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %s, want admin", claims.Role)
	}
}
