package api

import (
	"errors"
	"fmt"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type createUserRequest struct {
	Username  string `json:"username" binding:"required,alphanum"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type userResponse struct {
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	CreatedAt         time.Time `json:"created_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		Email:             user.Email,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		CreatedAt:         user.CreatedAt,
		PasswordChangedAt: user.PasswordChangedAt.Time,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		fmt.Printf("could not hash password: %v", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("An error occurred. "), http.StatusInternalServerError))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		HashedPassword: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.UniqueViolation {
			ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		}
		fmt.Printf("could not create user: %v", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("An error occurred. "), http.StatusInternalServerError))
		return
	}

	userResponse := newUserResponse(user)

	ctx.JSON(http.StatusCreated, successResponse(userResponse, http.StatusCreated))
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required"`
}

type loginUserResponse struct {
	Token                 string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

var ErrLoginFailed = errors.New("username/password is invalid")

func (server *Server) loginUser(ctx *gin.Context) {
	// Get the request body
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}
	// Get the user
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(ErrLoginFailed, http.StatusNotFound))
		return
	}
	// Compare passwords
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(ErrLoginFailed, http.StatusUnauthorized))
		return
	}

	// Generate token
	token, tokenPayload, err := server.tokenMaker.CreateToken(user.ID, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("An error occurred. "), http.StatusInternalServerError))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(user.ID, server.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("An error occurred. "), http.StatusInternalServerError))
		return
	}

	arg := db.CreateSessionParams{
		UserID:       user.ID,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.RemoteIP(),
		RefreshToken: refreshToken,
		ExpiresAt:    refreshPayload.ExpireAt,
		IsBlocked:    false,
	}

	_, err = server.store.CreateSession(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("An error occurred. "), http.StatusInternalServerError))
		return
	}

	// Return payload to user
	ctx.JSON(http.StatusOK, successResponse(loginUserResponse{
		Token:                 token,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  tokenPayload.ExpireAt,
		RefreshTokenExpiresAt: refreshPayload.ExpireAt,
		User:                  newUserResponse(user),
	}, http.StatusOK))

}
