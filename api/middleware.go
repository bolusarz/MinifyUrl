package api

import (
	"errors"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/token"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey  = "Authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("auth: authorization header is empty")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, http.StatusUnauthorized))
			return
		}

		fields := strings.Fields(authorizationHeader)

		if len(fields) != 2 {
			err := errors.New("auth: authorization header is invalid")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, http.StatusUnauthorized))
			return
		}

		authorizationType := strings.ToLower(fields[0])

		if authorizationType != authorizationTypeBearer {
			err := errors.New("auth: authorization type is invalid")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, http.StatusUnauthorized))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, http.StatusUnauthorized))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}

func userExistsMiddleware(store db.Store) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

		_, err := store.GetUserById(ctx, authPayload.UserID)
		if err != nil {
			err := errors.New("auth: invalid token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err, http.StatusUnauthorized))
		}

		ctx.Next()
	}
}
