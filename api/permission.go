package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
)

type createPermissionRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description" binding:"max=255"`
}

type permissionResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newPermissionResponse(permission db.Permission) permissionResponse {
	return permissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description.String,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func (server *Server) createPermission(ctx *gin.Context) {
	var req createPermissionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreatePermissionParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	}

	permission, err := server.store.CreatePermission(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newPermissionResponse(permission)
	ctx.JSON(http.StatusOK, successResponse("Permission created successfully", rsp))
}

type getPermissionRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getPermission(ctx *gin.Context) {
	var req getPermissionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	permission, err := server.store.GetPermission(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newPermissionResponse(permission)
	ctx.JSON(http.StatusOK, successResponse("Permission retrieved successfully", rsp))
}

type listPermissionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

type permissionsResponse struct {
	Meta struct {
		Page       int32 `json:"page"`
		TotalPages int32 `json:"total_pages"`
		TotalCount int64 `json:"total_count"`
	} `json:"meta"`
	Data []permissionResponse `json:"data"`
}

func (server *Server) listPermissions(ctx *gin.Context) {
	var req listPermissionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListPermissionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	permissions, err := server.store.ListPermissions(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalCount, err := server.store.CountPermissions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalPages := int32(totalCount) / req.PageSize
	if int32(totalCount)%req.PageSize != 0 {
		totalPages++
	}

	rsp := permissionsResponse{
		Meta: struct {
			Page       int32 `json:"page"`
			TotalPages int32 `json:"total_pages"`
			TotalCount int64 `json:"total_count"`
		}{
			Page:       req.PageID,
			TotalPages: totalPages,
			TotalCount: totalCount,
		},
		Data: make([]permissionResponse, len(permissions)),
	}
	for i, permission := range permissions {
		rsp.Data[i] = newPermissionResponse(permission)
	}

	ctx.JSON(http.StatusOK, successResponse("Permissions retrieved successfully", rsp))
}

type updatePermissionRequest struct {
	Name        string `json:"name" binding:"max=255"`
	Description string `json:"description" binding:"max=255"`
}

func (server *Server) updatePermission(ctx *gin.Context) {
	var reqURI getPermissionRequest
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var reqJSON updatePermissionRequest
	if err := ctx.ShouldBindJSON(&reqJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdatePermissionParams{
		ID: reqURI.ID,
		Name: sql.NullString{
			String: reqJSON.Name,
			Valid:  reqJSON.Name != "",
		},
		Description: sql.NullString{
			String: reqJSON.Description,
			Valid:  reqJSON.Description != "",
		},
	}

	permission, err := server.store.UpdatePermission(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newPermissionResponse(permission)
	ctx.JSON(http.StatusOK, successResponse("Permission updated successfully", rsp))
}

func (server *Server) deletePermission(ctx *gin.Context) {
	var req getPermissionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeletePermission(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Permission deleted successfully", nil))
}
