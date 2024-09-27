package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type renewAccessTokenPayload struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("refresh token not provided"), http.StatusBadRequest))
		return
	}

	payload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("invalid token"), http.StatusUnauthorized))
		return
	}

	session, err := server.store.GetSession(ctx, payload.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("no session created"), http.StatusBadRequest))
		return
	}

	if time.Now().After(session.ExpiresAt) {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("session expired"), http.StatusBadRequest))
		return
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("session expired"), http.StatusBadRequest))
		return
	}

	accessToken, accessTokenPayload, err := server.tokenMaker.CreateToken(session.UserID, server.config.AccessTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("an error occurred"), http.StatusInternalServerError))
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpireAt,
	}

	ctx.JSON(200, successResponse(rsp, 200))

}
