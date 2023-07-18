package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"gofaster"
	"net/http"
	"time"
)

const JWTToken = "gofaster_token"

type JwtHandler struct {
	Alg           string //jwt算法
	Authenticator func(ctx *gofaster.Context) (map[string]any, error)
	TimeOut       time.Duration
	TimeFunc      func() time.Time
	Key           []byte
	privateKey    string
	SendCookie    bool

	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	SecureCookie   bool
	CookieHttpOnly bool
	RefreshTimeOut time.Duration
	RefreshKey     string
	Header         string
	AuthHandler    func(ctx *gofaster.Context, err error)
}

type JwtResponse struct {
	Token        string
	RefreshToken string
}

// 登录，用户名密码认证->用户ID,生成token，保存在cookie中
func (h *JwtHandler) LoginHandler(ctx *gofaster.Context) (*JwtResponse, error) {
	data, err := h.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if h.Alg == "" {
		h.Alg = "HS256"
	}
	//A部分
	method := jwt.GetSigningMethod(h.Alg)
	token := jwt.New(method)
	//B部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}
	if h.TimeFunc == nil {
		h.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := h.TimeFunc().Add(h.TimeOut)
	//过期时间
	claims["exp"] = expire.Unix()
	//发布时间
	claims["iat"] = h.TimeFunc().Unix()
	//C部分
	var signedString string
	if h.usingPublicKeyAlgo() {
		signedString, err = token.SignedString(h.privateKey)
	} else {
		signedString, err = token.SignedString(h.Key)
	}
	if err != nil {
		return nil, err
	}
	//refreshToken
	refreshToken, err := h.refreshToken(token)
	if err != nil {
		return nil, err
	}
	jwtResponse := &JwtResponse{
		Token:        signedString,
		RefreshToken: refreshToken,
	}
	if h.SendCookie {
		if h.CookieName == "" {
			h.CookieName = JWTToken
		}
		if h.CookieMaxAge == 0 {
			h.CookieMaxAge = expire.Unix() - h.TimeFunc().Unix()
		}
		ctx.SetCookie(h.CookieName, signedString, h.CookieMaxAge, "/", h.CookieDomain, h.SecureCookie, h.CookieHttpOnly)
	}
	return jwtResponse, nil
}

func (h *JwtHandler) usingPublicKeyAlgo() bool {
	switch h.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (h *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = h.TimeFunc().Add(h.RefreshTimeOut).Unix()
	var signedString string
	var err error
	if h.usingPublicKeyAlgo() {
		signedString, err = token.SignedString(h.privateKey)
	} else {
		signedString, err = token.SignedString(h.Key)
	}
	if err != nil {
		return "", err
	}
	return signedString, nil
}

// 退出登录
func (h *JwtHandler) LogoutHandler(ctx *gofaster.Context) error {
	if h.SendCookie {
		if h.CookieName == "" {
			h.CookieName = JWTToken
		}
		ctx.SetCookie(h.CookieName, "", -1, "/", h.CookieDomain, h.SecureCookie, h.CookieHttpOnly)
		return nil
	}
	return nil
}

// 刷新token
func (h *JwtHandler) RefreshHandler(ctx *gofaster.Context) (*JwtResponse, error) {
	rToken, ok := ctx.Get(h.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}

	if h.Alg == "" {
		h.Alg = "HS256"
	}
	//解析token
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if h.usingPublicKeyAlgo() {
			return h.privateKey, nil
		} else {
			return h.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}

	//B部分
	claims := t.Claims.(jwt.MapClaims)

	if h.TimeFunc == nil {
		h.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	expire := h.TimeFunc().Add(h.TimeOut)
	//过期时间
	claims["exp"] = expire.Unix()
	//发布时间
	claims["iat"] = h.TimeFunc().Unix()
	//C部分
	var signedString string
	if h.usingPublicKeyAlgo() {
		signedString, err = t.SignedString(h.privateKey)
	} else {
		signedString, err = t.SignedString(h.Key)
	}
	if err != nil {
		return nil, err
	}
	//refreshToken
	refreshToken, err := h.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jwtResponse := &JwtResponse{
		Token:        signedString,
		RefreshToken: refreshToken,
	}
	if h.SendCookie {
		if h.CookieName == "" {
			h.CookieName = JWTToken
		}
		if h.CookieMaxAge == 0 {
			h.CookieMaxAge = expire.Unix() - h.TimeFunc().Unix()
		}
		ctx.SetCookie(h.CookieName, signedString, h.CookieMaxAge, "/", h.CookieDomain, h.SecureCookie, h.CookieHttpOnly)
	}
	return jwtResponse, nil
}

func (h *JwtHandler) AuthInterceptor(next gofaster.HandlerFunc) gofaster.HandlerFunc {
	return func(ctx *gofaster.Context) {
		if h.Header == "" {
			h.Header = "Authorization"
		}
		token := ctx.R.Header.Get(h.Header)
		if token == "" {
			cookie, err := ctx.R.Cookie(h.CookieName)
			if err != nil {
				h.AuthErrorHandler(ctx, err)
				return
			}
			token = cookie.String()
		}
		if token == "" {
			h.AuthErrorHandler(ctx, errors.New("token is null"))
			return
		}
		//解析token
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if h.usingPublicKeyAlgo() {
				return h.privateKey, nil
			} else {
				return h.Key, nil
			}
		})
		if err != nil {
			h.AuthErrorHandler(ctx, err)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		ctx.Set("jwt_claims", claims)
		next(ctx)
	}
}

func (h *JwtHandler) AuthErrorHandler(ctx *gofaster.Context, err error) {
	if h.AuthHandler == nil {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	} else {
		h.AuthHandler(ctx, nil)
	}
}
