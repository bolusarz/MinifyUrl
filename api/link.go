package api

import (
	"errors"
	db "github.com/bolusarz/urlmini/db/sqlc"
	"github.com/bolusarz/urlmini/token"
	"github.com/bolusarz/urlmini/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

type createLinkParams struct {
	Link string `json:"link" binding:"required"`
	Code string `json:"code"`
}

func (server *Server) CreateLink(ctx *gin.Context) {
	var req createLinkParams

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	if req.Code == "" {
		req.Code = util.RandomCode()
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateLinkParams{
		Code:   req.Code,
		Link:   req.Link,
		UserID: authPayload.UserID,
	}

	link, err := server.store.CreateLink(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
	}

	ctx.JSON(http.StatusCreated, successResponse(link, http.StatusCreated))

}

type getLinkByCodeParams struct {
	Code string `uri:"code" binding:"required,alpha"`
}

func (server *Server) GetLinkByCode(ctx *gin.Context) {
	var req getLinkByCodeParams

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	link, err := server.store.GetLinkByCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err, http.StatusNotFound))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	if !link.Active.Bool {
		ctx.JSON(http.StatusNotFound, errorResponse(err, http.StatusNotFound))
		return
	}

	ctx.Redirect(http.StatusPermanentRedirect, link.Link)

}

type getLinkByIDParams struct {
	ID int64 `uri:"id" binding:"required,number,min=1"`
}

func (server *Server) GetLinkById(ctx *gin.Context) {
	var req getLinkByIDParams

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	link, err := server.store.GetLinkById(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err, http.StatusNotFound))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	if link.UserID != authPayload.UserID {
		ctx.JSON(http.StatusForbidden, errorResponse(err, http.StatusForbidden))
		return
	}

	ctx.JSON(http.StatusOK, successResponse(link, http.StatusOK))
	return
}

type getLinksParams struct {
	PageID   int32 `form:"page_id"`
	PageSize int32 `form:"page_size"`
}

// TODO: Add the previous page id and the nextpage id
func (server *Server) GetLinks(ctx *gin.Context) {
	var req getLinksParams

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	if req.PageID == 0 {
		req.PageID = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	args := db.GetLinksByUserParams{
		UserID: authPayload.UserID,
		Offset: (req.PageID - 1) * req.PageSize,
		Limit:  req.PageSize,
	}

	links, err := server.store.GetLinksByUser(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	ctx.JSON(http.StatusOK, successResponse(links, http.StatusOK))
	return
}

type toggleLinkStatusParams struct {
	ID int64 `uri:"id" binding:"required,number,min=1"`
}

func (server *Server) ToggleLinkStatus(ctx *gin.Context) {
	var req toggleLinkStatusParams

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	link, err := server.store.GetLinkById(ctx, req.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err, http.StatusNotFound))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if link.UserID != authPayload.UserID {
		ctx.JSON(http.StatusForbidden, errorResponse(err, http.StatusForbidden))
		return
	}

	args := db.ToggleStatusParams{
		ID: link.ID,
	}

	link, err = server.store.ToggleStatus(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	ctx.JSON(http.StatusOK, successResponse(link, http.StatusOK))
	return
}

type changeCodeParams struct {
	Code string `json:"code" binding:"required,alpha"`
}

func (server *Server) ChangeCode(ctx *gin.Context) {
	var req changeCodeParams
	var linkReq getLinkByIDParams

	if err := ctx.ShouldBindUri(&linkReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err, http.StatusBadRequest))
		return
	}

	link, err := server.store.GetLinkById(ctx, linkReq.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err, http.StatusNotFound))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if link.UserID != authPayload.UserID {
		ctx.JSON(http.StatusForbidden, errorResponse(err, http.StatusForbidden))
		return
	}

	args := db.UpdateCodeParams{
		ID:   link.ID,
		Code: req.Code,
	}

	link, err = server.store.UpdateCode(ctx, args)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err, http.StatusInternalServerError))
		return
	}

	ctx.JSON(http.StatusOK, successResponse(link, http.StatusOK))
	return
}
