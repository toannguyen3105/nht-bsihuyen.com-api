package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
)

type createRolePermissionRequest struct {
	RoleID       int32 `json:"role_id" binding:"required,min=1"`
	PermissionID int32 `json:"permission_id" binding:"required,min=1"`
}

func (server *Server) createRolePermission(ctx *gin.Context) {
	var req createRolePermissionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateRolePermissionParams{
		RoleID:       req.RoleID,
		PermissionID: req.PermissionID,
	}

	rolePermission, err := server.store.CreateRolePermission(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, rolePermission)
}

type deleteRolePermissionRequest struct {
	RoleID       int32 `uri:"role_id" binding:"required,min=1"`
	PermissionID int32 `uri:"permission_id" binding:"required,min=1"`
}

func (server *Server) deleteRolePermission(ctx *gin.Context) {
	var req deleteRolePermissionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.DeleteRolePermissionParams{
		RoleID:       req.RoleID,
		PermissionID: req.PermissionID,
	}

	err := server.store.DeleteRolePermission(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Role permission deleted successfully", nil))
}

type listRolePermissionsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

type rolePermissionsResponse struct {
	Meta struct {
		Page       int32 `json:"page"`
		TotalPages int32 `json:"total_pages"`
		TotalCount int64 `json:"total_count"`
	} `json:"meta"`
	Data []db.RolePermission `json:"data"`
}

func (server *Server) listRolePermissions(ctx *gin.Context) {
	var req listRolePermissionsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListRolePermissionsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	rolePermissions, err := server.store.ListRolePermissions(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalCount, err := server.store.CountRolePermissions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalPages := int32(totalCount) / req.PageSize
	if int32(totalCount)%req.PageSize != 0 {
		totalPages++
	}

	rsp := rolePermissionsResponse{
		Meta: struct {
			Page       int32 `json:"page"`
			TotalPages int32 `json:"total_pages"`
			TotalCount int64 `json:"total_count"`
		}{
			Page:       req.PageID,
			TotalPages: totalPages,
			TotalCount: totalCount,
		},
		Data: rolePermissions,
	}

	ctx.JSON(http.StatusOK, successResponse("Role permissions retrieved successfully", rsp))
}

func (server *Server) getRolePermission(ctx *gin.Context) {
	var req deleteRolePermissionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetRolePermissionParams{
		RoleID:       req.RoleID,
		PermissionID: req.PermissionID,
	}

	rolePermission, err := server.store.GetRolePermission(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Role permission retrieved successfully", rolePermission))
}

type updateRolePermissionRequest struct {
	PermissionID int32 `json:"permission_id" binding:"required,min=1"`
}

func (server *Server) updateRolePermission(ctx *gin.Context) {
	var reqURI deleteRolePermissionRequest
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var reqJSON updateRolePermissionRequest
	if err := ctx.ShouldBindJSON(&reqJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateRolePermissionParams{
		RoleID:         reqURI.RoleID,
		PermissionID:   reqURI.PermissionID,
		PermissionID_2: reqJSON.PermissionID,
	}

	rolePermission, err := server.store.UpdateRolePermission(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Role permission updated successfully", rolePermission))
}
