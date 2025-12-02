package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
)

type createMedicineRequest struct {
	Name        string  `json:"name" binding:"required"`
	Unit        string  `json:"unit" binding:"required,oneof=tablet capsule box bottle"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Stock       int32   `json:"stock" binding:"required,min=0"`
	Description *string `json:"description"`
}

func (server *Server) createMedicine(ctx *gin.Context) {
	var req createMedicineRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateMedicineParams{
		Name:  req.Name,
		Unit:  req.Unit,
		Price: fmt.Sprintf("%.2f", req.Price),
		Stock: req.Stock,
	}

	if req.Description != nil {
		arg.Description = sql.NullString{String: *req.Description, Valid: true}
	}

	medicine, err := server.store.CreateMedicine(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Medicine created successfully", medicine))
}

type getMedicineRequest struct {
	ID int32 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getMedicine(ctx *gin.Context) {
	var req getMedicineRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	medicine, err := server.store.GetMedicine(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Medicine retrieved successfully", medicine))
}

type listMedicinesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

type medicinesResponse struct {
	Meta struct {
		Page       int32 `json:"page"`
		TotalPages int32 `json:"total_pages"`
		TotalCount int64 `json:"total_count"`
	} `json:"meta"`
	Data []db.Medicine `json:"data"`
}

func (server *Server) listMedicines(ctx *gin.Context) {
	var req listMedicinesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListMedicinesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	medicines, err := server.store.ListMedicines(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalCount, err := server.store.CountMedicines(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	totalPages := int32(totalCount) / req.PageSize
	if int32(totalCount)%req.PageSize != 0 {
		totalPages++
	}

	rsp := medicinesResponse{
		Meta: struct {
			Page       int32 `json:"page"`
			TotalPages int32 `json:"total_pages"`
			TotalCount int64 `json:"total_count"`
		}{
			Page:       req.PageID,
			TotalPages: totalPages,
			TotalCount: totalCount,
		},
		Data: medicines,
	}

	ctx.JSON(http.StatusOK, successResponse("Medicines retrieved successfully", rsp))
}

type updateMedicineRequest struct {
	Name        *string  `json:"name"`
	Unit        *string  `json:"unit" binding:"omitempty,oneof=tablet capsule box bottle"`
	Price       *float64 `json:"price" binding:"omitempty,min=0"`
	Stock       *int32   `json:"stock" binding:"omitempty,min=0"`
	Description *string  `json:"description"`
}

func (server *Server) updateMedicine(ctx *gin.Context) {
	var reqUri getMedicineRequest
	if err := ctx.ShouldBindUri(&reqUri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var reqBody updateMedicineRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateMedicineParams{
		ID: reqUri.ID,
	}

	if reqBody.Name != nil {
		arg.Name = sql.NullString{String: *reqBody.Name, Valid: true}
	}
	if reqBody.Unit != nil {
		arg.Unit = *reqBody.Unit
	}
	if reqBody.Price != nil {
		arg.Price = sql.NullString{String: fmt.Sprintf("%.2f", *reqBody.Price), Valid: true}
	}
	if reqBody.Stock != nil {
		arg.Stock = sql.NullInt32{Int32: *reqBody.Stock, Valid: true}
	}
	if reqBody.Description != nil {
		arg.Description = sql.NullString{String: *reqBody.Description, Valid: true}
	}

	medicine, err := server.store.UpdateMedicine(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Medicine updated successfully", medicine))
}

func (server *Server) deleteMedicine(ctx *gin.Context) {
	var req getMedicineRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteMedicine(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, successResponse("Medicine deleted successfully", nil))
}
