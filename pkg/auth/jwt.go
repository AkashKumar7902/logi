package auth

import (
    "errors"
    "time"

    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    jwtSecret []byte
}

func NewAuthService(secret string) *AuthService {
    return &AuthService{
        jwtSecret: []byte(secret),
    }
}

func (a *AuthService) GenerateJWT(userID string, role string) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "exp":  time.Now().Add(72 * time.Hour).Unix(),
        "sub":  userID,
        "role": role,
    })
    return token.SignedString(a.jwtSecret)
}

func (a *AuthService) ValidateJWT(tokenString string) (string, string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return a.jwtSecret, nil
    })
    if err != nil || !token.Valid {
        return "", "", errors.New("invalid token")
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return "", "", errors.New("invalid token claims")
    }
    userID, ok := claims["sub"].(string)
    if !ok {
        return "", "", errors.New("invalid token subject")
    }
    role, ok := claims["role"].(string)
    if !ok {
        return "", "", errors.New("invalid token role")
    }
    return userID, role, nil
}

func (a *AuthService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func (a *AuthService) CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
