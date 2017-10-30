package auth

import (
  	"github.com/dgrijalva/jwt-go"
	"github.com/adriamb/bcdapp/config"
	"time"
	"fmt"
)

func JwtCreateToken(address string) (string,error) {
	expiration := time.Now().Add(time.Hour*2)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	    // Subject
	    "sub": address,
	    // Expiration time
	    "exp": expiration.Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.C.ServerSecret))
	if err != nil {
		return "",err
	}

	return tokenString,nil
}

func JwtVerifyToken(tokenString string) (string,error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	        return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	    }
	    return []byte(config.C.ServerSecret), nil
	})	
	if err != nil {
		return "",err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
	    return claims["sub"].(string), nil
	} else {
	    return "",err
	}
}

