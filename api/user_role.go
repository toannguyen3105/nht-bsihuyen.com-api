package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
)

type userRoleRequest struct {
	UserID int32 `json:"user_id" binding:"required,min=1"`
	RoleID int32 `json:"role_id" binding:"required,min=1"`
}

func (server *Server) addUserRole(ctx *gin.Context) {
	var req userRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.AddRoleForUserParams{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}

	userRole, err := server.store.AddRoleForUser(ctx, arg)
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

	ctx.JSON(http.StatusOK, successResponse("User role created successfully", userRole))
}

type getUserRolesRequest struct {
	UserID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getUserRoles(ctx *gin.Context) {
	var req getUserRolesRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	roles, err := server.store.GetRolesForUser(ctx, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := make([]roleResponse, len(roles))
	for i, role := range roles {
		rsp[i] = newRoleResponse(role)
	}

	ctx.JSON(http.StatusOK, successResponse("User roles retrieved successfully", rsp))
}

func (server *Server) deleteUserRole(ctx *gin.Context) {
	var req userRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.RemoveRoleForUserParams{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}

	err := server.store.RemoveRoleForUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("User role deleted successfully", nil))
}

type listUserRolesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=20"`
}

func (server *Server) listUserRoles(ctx *gin.Context) {
	var req listUserRolesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUserRolesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	userRoles, err := server.store.ListUserRoles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("User roles retrieved successfully", userRoles))
}