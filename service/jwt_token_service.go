package service

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt"
)

type UserClaims struct {
	UserID    string `json:"userID"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email			string `json:"email"`
	jwt.StandardClaims
}

func NewAccessToken(claims UserClaims) (string, error) {
	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	keyData, err := os.ReadFile("keys/token_key.key")
	if err != nil {
		println("Error reading file")
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		println("Error parsing key")
		return "", err
	}

	return newAccessToken.SignedString(key)
}

func NewRefreshToken(claims jwt.StandardClaims) (string, error) {
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	keyData, err := os.ReadFile("keys/token_key.key")
	if err != nil {
		println("Error reading file")
		return "", err
	}
	
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		println("Error parsing key")
		return "", err
	}

	return newRefreshToken.SignedString(key)
}

func ParseAccessToken(accessToken string) *UserClaims {
	f, err := os.ReadFile("keys/token_key.key")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(f)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.PublicKey

	claims := &UserClaims{}

	decodedToken, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &publicKey, nil
	});

	if err != nil {
		println("error parsing access token")
		return nil
	}

	// Check if the token is valid and claims are of type *UserClaims
	if claims, ok := decodedToken.Claims.(*UserClaims); ok && decodedToken.Valid {
		return claims
	} else {
		println("error with access claims")
		return nil
	}
}

func ParseRefreshToken(refreshToken string) *jwt.StandardClaims {
	f, err := os.ReadFile("keys/token_key.key")
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(f)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.PublicKey

	claims := &jwt.StandardClaims{}

	decodedToken, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return &publicKey, nil
	});

	if err != nil {
		println("error parsing refresh token")
		return nil
	}

	// Check if the token is valid and claims are of type *UserClaims
	if claims, ok := decodedToken.Claims.(*jwt.StandardClaims); ok && decodedToken.Valid {
		return claims
	} else {
		println("error with refresh claims")
		return nil
	}
}
