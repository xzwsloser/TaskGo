package middleware

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/xzwsloser/TaskGo/admin/internal/model/resp"
	"github.com/xzwsloser/TaskGo/pkg/logger"
	"golang.org/x/sync/singleflight"
)

/*
	@Description: Create & Parse Jwt Token
*/
// Custom claims structure
type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

type BaseClaims struct {
	ID       int
	UserName string
}

type JWT struct {
	SigningKey []byte
}

var (
	TokenExpired     = errors.New("Token is expired")
	TokenNotValidYet = errors.New("Token not active yet")
	TokenMalformed   = errors.New("That's not even a token")
	TokenInvalid     = errors.New("Couldn't handle this token")
)

var control = &singleflight.Group{}

func NewJWT() *JWT {
	return &JWT{
		[]byte("S0dEdN9tqG0AAAAHdElNRQfmCgwBDCSd2zTMAAAA"),
	}
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	now := time.Now().Unix()
	claims := CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: 86400, // buffer time 1 day buffer time will get a new token refresh token. 
		RegisteredClaims: jwt.RegisteredClaims{
			// jwt token 生效事件
			NotBefore: jwt.NewNumericDate(time.Unix(now-1000,0)),
			// jwt token 过期事件(7天)
			ExpiresAt: jwt.NewNumericDate(time.Unix(now+604800,0)),
			// jwt 签发时间
			IssuedAt: jwt.NewNumericDate(time.Unix(now, 0)),
			// jwt 签发者
			Issuer:    "xzw",                    // the publisher of the signature
		},
	}
	return claims
}

// create a token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// Replacing old token with new token using merging and origin-pull to avoid concurrency problems
func (j *JWT) CreateTokenByOldToken(oldToken string, claims CustomClaims) (string, error) {
	// Only Operation The Key For Once
	v, err, _ := control.Do("JWT:"+oldToken, func() (any, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		return j.SigningKey, nil
	})
	if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, TokenMalformed
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, TokenExpired
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, TokenNotValidYet
		} else {
			return nil, TokenInvalid
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid
	} else {
		return nil, TokenInvalid
	}
}

func GetClaims(c *gin.Context) (*CustomClaims, error) {
	token := c.Request.Header.Get("Authorization")
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		logger.GetLogger().Error("Failed to obtain parsing information from jwt from Context of Gin. Please check whether Authorization exists in the request header and whether claims is the specified structure.")
	}
	return claims, err
}

// Get the user roles parsed from jwt from the Context of Gin
func GetUserInfo(c *gin.Context) *CustomClaims {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return nil
		} else {
			return cl
		}
	} else {
		waitUse := claims.(*CustomClaims)
		return waitUse
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// We have jwt authentication header information to return token information when Authorization logs in. Here,
		//the front end needs to store the token in cookie or local localStorage,
		//but you need to negotiate the expiration time with the back end.
		//You can agree to refresh the token or log in again.
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, "未登录或非法访问", c)
			c.Abort()
			return
		}
		j := NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == TokenExpired {
				resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, "授权已过期", c)
				c.Abort()
				return
			}
			resp.FailWithDetailed(resp.ERROR, gin.H{"reload": true}, err.Error(), c)
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
