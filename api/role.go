package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
)

type createRoleRequest struct {
	Name        string `json:"name" binding:"required,alphanum,max=255"`
	Description string `json:"description" binding:"max=255"`
}

type roleResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newRoleResponse(role db.Role) roleResponse {
	return roleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description.String,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func (server *Server) createRole(ctx *gin.Context) {
	var req createRoleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateRoleParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	}

	role, err := server.store.CreateRole(ctx, arg)
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

	rsp := newRoleResponse(role)
	ctx.JSON(http.StatusOK, successResponse("Role created successfully", rsp))
}

type getRoleRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getRole(ctx *gin.Context) {
	var req getRoleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	role, err := server.store.GetRole(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newRoleResponse(role)
	ctx.JSON(http.StatusOK, successResponse("Role retrieved successfully", rsp))
}

type listRolesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=10,max=100"`
}

type rolesResponse struct {
	Meta struct {
		Page       int32 `json:"page"`
		TotalPages int32 `json:"total_pages"`
		TotalCount int64 `json:"total_count"`
	} `json:"meta"`
	Data []roleResponse `json:"data"`
}

func (server *Server) listRoles(ctx *gin.Context) {
	var req listRolesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListRolesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	roles, err := server.store.ListRoles(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalCount, err := server.store.CountRoles(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalPages := int32(totalCount) / req.PageSize
	if int32(totalCount)%req.PageSize != 0 {
		totalPages++
	}

	rsp := rolesResponse{
		Meta: struct {
			Page       int32 `json:"page"`
			TotalPages int32 `json:"total_pages"`
			TotalCount int64 `json:"total_count"`
		}{
			Page:       req.PageID,
			TotalPages: totalPages,
			TotalCount: totalCount,
		},
		Data: make([]roleResponse, len(roles)),
	}
	for i, role := range roles {
		rsp.Data[i] = newRoleResponse(role)
	}

	ctx.JSON(http.StatusOK, successResponse("Roles retrieved successfully", rsp))
}

type updateRoleRequest struct {
	Name        string `json:"name" binding:"alphanum,max=255"`
	Description string `json:"description" binding:"max=255"`
}

func (server *Server) updateRole(ctx *gin.Context) {
	var reqURI getRoleRequest
	if err := ctx.ShouldBindUri(&reqURI); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var reqJSON updateRoleRequest
	if err := ctx.ShouldBindJSON(&reqJSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateRoleParams{
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

	role, err := server.store.UpdateRole(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newRoleResponse(role)
	ctx.JSON(http.StatusOK, successResponse("Role updated successfully", rsp))
}

func (server *Server) deleteRole(ctx *gin.Context) {
	var req getRoleRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteRole(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Role deleted successfully", nil))
}
