package api

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/token"
	"github.com/toannguyen3105/nht-bsihuyen.com-api/utils"
)

type Server struct {
	config     utils.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://your-frontend.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.GET("/users", server.requirePermission("VIEW_SCREEN_USER"), server.listUsers)

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.listAccount)

	authRoutes.POST("/transfers", server.createTransfer)

	authRoutes.POST("/roles", server.requirePermission("VIEW_SCREEN_ROLE"), server.createRole)
	authRoutes.GET("/roles", server.requirePermission("VIEW_SCREEN_ROLE"), server.listRoles)
	authRoutes.GET("/roles/:id", server.requirePermission("VIEW_SCREEN_ROLE"), server.getRole)
	authRoutes.PUT("/roles/:id", server.requirePermission("VIEW_SCREEN_ROLE"), server.updateRole)
	authRoutes.DELETE("/roles/:id", server.requirePermission("VIEW_SCREEN_ROLE"), server.deleteRole)

	authRoutes.POST("/permissions", server.requirePermission("VIEW_SCREEN_PERMISSION"), server.createPermission)
	authRoutes.GET("/permissions/:id", server.requirePermission("VIEW_SCREEN_PERMISSION"), server.getPermission)
	authRoutes.GET("/permissions", server.requirePermission("VIEW_SCREEN_PERMISSION"), server.listPermissions)
	authRoutes.PUT("/permissions/:id", server.requirePermission("VIEW_SCREEN_PERMISSION"), server.updatePermission)
	authRoutes.DELETE("/permissions/:id", server.requirePermission("VIEW_SCREEN_PERMISSION"), server.deletePermission)

	authRoutes.POST("/role_permissions", server.requirePermission("VIEW_SCREEN_ROLE_PERMISSION"), server.createRolePermission)
	authRoutes.GET("/role_permissions", server.requirePermission("VIEW_SCREEN_ROLE_PERMISSION"), server.listRolePermissions)
	authRoutes.GET("/role_permissions/:role_id/:permission_id", server.requirePermission("VIEW_SCREEN_ROLE_PERMISSION"), server.getRolePermission)
	authRoutes.PUT("/role_permissions/:role_id/:permission_id", server.requirePermission("VIEW_SCREEN_ROLE_PERMISSION"), server.updateRolePermission)
	authRoutes.DELETE("/role_permissions/:role_id/:permission_id", server.requirePermission("VIEW_SCREEN_ROLE_PERMISSION"), server.deleteRolePermission)

	authRoutes.POST("/user-roles", server.requirePermission("VIEW_SCREEN_USER_ROLE"), server.addUserRole)
	authRoutes.GET("/users/:id/roles", server.requirePermission("VIEW_SCREEN_USER_ROLE"), server.getUserRoles)
	authRoutes.DELETE("/user-roles", server.requirePermission("VIEW_SCREEN_USER_ROLE"), server.deleteUserRole)
	authRoutes.PUT("/user-roles", server.requirePermission("VIEW_SCREEN_USER_ROLE"), server.updateUserRole)
	authRoutes.GET("/user-roles", server.requirePermission("VIEW_SCREEN_USER_ROLE"), server.listUserRoles)

	// For testing
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func successResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

func errorResponse(err error) APIResponse {
	return APIResponse{
		Status:  "error",
		Message: err.Error(),
		Data:    nil,
	}
}
