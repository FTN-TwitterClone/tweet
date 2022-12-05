package jwt

import (
	"context"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"os"
	"time"
	"tweet/model"
)

func ExtractJWTUserMiddleware(tracer trace.Tracer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newCtx, span := tracer.Start(r.Context(), "ExtractJWTUserMiddleware")
			defer span.End()

			if authHeader, ok := r.Header["Authorization"]; ok {
				tokenString := authHeader[0]

				_, parseSpan := tracer.Start(newCtx, "jwt.Parse")
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("SECRET_KEY")), nil
				})
				parseSpan.End()

				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					authUser := model.AuthUser{
						Username: claims["username"].(string),
						Role:     claims["role"].(string),
						Exp:      time.UnixMilli(int64(claims["exp"].(float64))),
					}

					span.SetAttributes(attribute.String("user", authUser.Username))

					authCtx := context.WithValue(newCtx, "authUser", authUser)

					next.ServeHTTP(w, r.WithContext(authCtx))
				} else {
					span.SetStatus(codes.Error, err.Error())
					http.Error(w, "Invalid token", 403)
				}
			} else {
				http.Error(w, "Invalid token", 403)
			}
		})
	}
}
