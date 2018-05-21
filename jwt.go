package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type CustomClaims struct {
	ID int `json:"id"`
	//Name  string `json:"name"`
	//Email string `json:"email"`
	jwt.StandardClaims
}

type JWT struct {
	SigningKey []byte
}

type NbHandlerFunc struct {
	gin.HandlerFunc
}

var (
	TokenExpired     error  = errors.New("Token 过期")
	TokenNotValidYet error  = errors.New("Token 未激活")
	TokenMalformed   error  = errors.New("请提供访问Token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "yishion"
	//Claims           CustomClaims
)

func JWTLogin() gin.HandlerFunc {
	return func(c *gin.Context) {

		u := c.Query("u")
		p := c.Query("p")
		if u == "duan" && p == "yishion" {

			j := NewJWT()
			claims := CustomClaims{
				1,
				//		"duan",
				//		"duan@isnb.net",
				jwt.StandardClaims{
					ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), //time.Now().Add(24 * time.Hour).Unix()
					Issuer:    "Yishion",
				},
			}
			token, err := j.CreateToken(claims)

			if err != nil {
				c.String(http.StatusOK, err.Error())
				c.Abort()
			}
			c.Header("Authorization", "Bear "+token)
			c.String(http.StatusOK, token)
		}
	}
}

func JWTAuth() gin.HandlerFunc {

	return func(c *gin.Context) {
		token := c.DefaultQuery("token", "")

		if token == "" {
			token = c.Request.Header.Get("Authorization")
			if s := strings.Split(token, " "); len(s) == 2 {
				token = s[1]
			}
		}

		j := NewJWT()
		claims, err := j.ParseToken(token)

		if err != nil {
			if err == TokenExpired {
				if token, err = j.RefreshToken(token); err == nil {
					c.Header("Authorization", "Bear "+token)
					c.JSON(http.StatusOK, gin.H{"error": 0, "message": "refresh token", "token": token})
					return
				}
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": 1, "message": err.Error()})
			return
		}
		c.Set("claims", claims)
	}
}

func (h *NbHandlerFunc) IsAuthorization(c *gin.Context) bool {
	_, exists := c.Get("claims")
	return exists
}

//添加认证逻辑
func Auth(f1 func(*gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAuthorization(c) {
			f1(c)
		}
	}

}

func IsAuthorization(c *gin.Context) bool {
	_, exists := c.Get("claims")
	return exists
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

func GetSignKey() string {
	return SignKey
}

func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, TokenInvalid
}

func (j *JWT) RefreshToken(tokenString string) (string, error) {
	jwt.TimeFunc = func() time.Time {
		return time.Unix(0, 0)
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now

		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}

	return "", TokenInvalid
}

func (j *JWT) CreateToken(claims CustomClaims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.SigningKey)
}
