package api

import (
	"fmt"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/token"
	"github.com/bolusarz/urlmini/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Server struct {
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	config     util.Config
}

func NewServer(store db.Store, config util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/login", server.loginUser)
	router.POST("/token/refresh", server.renewAccessToken)

	router.GET("/:code", server.GetLinkByCode)

	authRoutes := router.Group("/").
		Use(
			authMiddleware(server.tokenMaker),
			userExistsMiddleware(server.store),
		)

	authRoutes.POST("/links", server.CreateLink)
	authRoutes.GET("/links", server.GetLinks)
	authRoutes.GET("/links/:id", server.GetLinkById)
	authRoutes.PATCH("/links/:id", server.ChangeCode)
	authRoutes.PATCH("/links/:id/toggle", server.ToggleLinkStatus)

	server.router = router
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

type Response[T any] struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  error  `json:"error"`
	Data   T      `json:"data"`
}

func errorResponse(err error, statusCode int) gin.H {
	return gin.H{
		"code":   statusCode,
		"status": http.StatusText(statusCode),
		"error":  err,
		"data":   nil,
	}
}

func successResponse(data interface{}, statusCode int) gin.H {
	return gin.H{
		"code":   statusCode,
		"status": http.StatusText(statusCode),
		"data":   data,
		"error":  nil,
	}
}
