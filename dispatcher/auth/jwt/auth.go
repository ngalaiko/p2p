package jwt

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/ngalayko/p2p/instance/peers"
	"github.com/ngalayko/p2p/logger"
)

const cookieName = "auth"

// Auth stores peer as a jwt token in cookies.
type Auth struct {
	logger *logger.Logger
	secret []byte
}

// New is a mock auth constructor.
func New(
	log *logger.Logger,
	secret string,
) *Auth {
	return &Auth{
		logger: log.Prefix("jwt"),
		secret: []byte(secret),
	}
}

type customClaims struct {
	*jwt.StandardClaims

	Peer *peers.Peer
}

// Store implements Authorizer.
func (a *Auth) Store(w http.ResponseWriter, p *peers.Peer) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &customClaims{
		Peer: p,
		StandardClaims: &jwt.StandardClaims{
			Subject:  p.ID,
			IssuedAt: time.Now().Unix(),
		},
	})

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return fmt.Errorf("error singing jwt token: %s", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  cookieName,
		Value: tokenString,
	})

	a.logger.Debug("stored peer %s in %s", p.ID, cookieName)

	return nil
}

// Get implements Authorizer.
func (a *Auth) Get(r *http.Request) (*peers.Peer, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		a.logger.Debug("%s is empty", cookieName)
		return nil, fmt.Errorf("can't get cookie: %s", err)
	}

	token, err := jwt.ParseWithClaims(cookie.Value, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims: %v", token.Claims)
	}

	a.logger.Debug("got peer %s from %s", claims.Peer.ID, cookieName)

	return claims.Peer, nil
}
