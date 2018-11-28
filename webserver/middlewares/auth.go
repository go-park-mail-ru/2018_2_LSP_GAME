package middlewares

import (
	cnt "context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2018_2_LSP_AUTH_GRPC/auth_proto"
	"github.com/go-park-mail-ru/2018_2_LSP_GAME/webserver/handlers"
	"github.com/gorilla/context"
)

// Auth Middleware for protecting urls from unauthorized users
func Auth(next handlers.HandlerFunc) handlers.HandlerFunc {
	return func(env *handlers.Env, w http.ResponseWriter, r *http.Request) error {
		signature, err := r.Cookie("signature")
		if err != nil {
			return handlers.StatusData{
				Code: http.StatusUnauthorized,
				Data: map[string]string{
					"error": "No signature cookie found",
				},
			}
		}

		headerPayload, err := r.Cookie("header.payload")
		if err != nil {
			return handlers.StatusData{
				Code: http.StatusUnauthorized,
				Data: map[string]string{
					"error": "No headerPayload cookie found",
				},
			}
		}

		tokenString := headerPayload.Value + "." + signature.Value

		ctx := cnt.Background()
		authManager := auth_proto.NewAuthCheckerClient(env.GRCPAuth)
		token, err := authManager.Check(ctx,
			&auth_proto.Token{
				Token: tokenString,
			})

		if err != nil {
			env.Logger.Fatalw("Error during grpc request",
				"err", err.Error(),
				"grpc", "user",
			)
			return handlers.StatusData{
				Code: http.StatusInternalServerError,
				Data: map[string]string{
					"error": "Internal server error",
				},
			}
		}

		if !token.Valid {
			return handlers.StatusData{
				Code: http.StatusUnauthorized,
				Data: map[string]string{
					"error": "Token is not valid",
				},
			}
		}

		firstDot := strings.Index(tokenString, ".") + 1
		secondDot := strings.Index(tokenString[firstDot:], ".") + firstDot
		claimsJSON, err := base64.StdEncoding.DecodeString(tokenString[firstDot:secondDot])

		if err != nil {
			env.Logger.Fatalw("Error during base64 string decoding",
				"err", err.Error(),
				"base64encoded", tokenString[firstDot:secondDot],
			)
			return handlers.StatusData{
				Code: http.StatusUnauthorized,
				Data: map[string]string{
					"error": "Internal server error",
				},
			}
		}

		claims := make(map[string]interface{})
		err = json.Unmarshal(claimsJSON, &claims)
		if err != nil {
			env.Logger.Warnw("Can't unmarshall data",
				"err", err.Error(),
				"data", claims,
				"json", claimsJSON,
			)
			return handlers.StatusData{
				Code: http.StatusUnauthorized,
				Data: map[string]string{
					"error": "Token is not valid",
				},
			}
		}

		context.Set(r, "claims", claims)

		return next(env, w, r)
	}
}
