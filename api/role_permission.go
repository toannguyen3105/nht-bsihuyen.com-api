package api

import (
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

	// TODO: Add authorization check here. Only admins should be able to assign permissions.

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
