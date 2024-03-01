package middleware
// User represents a user with a username, password, and role.
import (
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	"net/http"
	"time"
	"strings"
	"fmt"
)

type User struct {
	Username string
	Password string
	Role     string
}

// Users is a mock database of users.
var users = []User{
	{Username: "mazda", Password: "Test1ng!", Role: "admin"},
	{Username: "honda", Password: "Test1ng!", Role: "editor"},
	{Username: "kia", Password: "Test1ng!", Role: "editor"},
	{Username: "benz", Password: "Test1ng!", Role: "viewer"},
	{Username: "toyota", Password: "Test1ng!", Role: "viewer"},
}

var jwtKey = []byte("JAIJAFFA")

// Credentials are used for parsing login requests.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims are used for creating JWT tokens.
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}


func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Authenticate the user
	var userRole string
	for _, user := range users {
		if user.Username == creds.Username && user.Password == creds.Password {
			userRole = user.Role
			break
		}
	}

	if userRole == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		Role:     userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// http.SetCookie(w, &http.Cookie{
	// 	Name:    "token",
	// 	Value:   tokenString,
	// 	Expires: expirationTime,
	// })
	// Instead of setting the token as a cookie, return it in the response body
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}


// IsAuthorized checks if the request is made by an authorized and authenticated user with the correct role
func IsAuthorized(requiredRole string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Expecting "Bearer <token>"
			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 {
				http.Error(w, "Invalid Authorization token format", http.StatusUnauthorized)
				return
			}

			tokenString := bearerToken[1]
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					http.Error(w, "Invalid token signature", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Check if the role in the token matches the required role
			if claims.Role != requiredRole {
				msg := fmt.Sprintf("Insufficient permissions: require role %s, but got role %s", requiredRole, claims.Role)
				http.Error(w, msg, http.StatusForbidden)
				return
			}

			// User is authorized; proceed with the next handler
			next.ServeHTTP(w, r)
		}
	}
}