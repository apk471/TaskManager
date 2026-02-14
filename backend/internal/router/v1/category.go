package v1

import (
	"github.com/apk471/go-taskmanager/internal/handler"
	"github.com/apk471/go-taskmanager/internal/middleware"
	"github.com/labstack/echo/v4"
)

func registerCategoryRoutes(r *echo.Group, h *handler.CategoryHandler, auth *middleware.AuthMiddleware) {
	// Category operations
	categories := r.Group("/categories")
	categories.Use(auth.RequireAuth)

	// Category collection operations
	categories.POST("", h.CreateCategory)
	categories.GET("", h.GetCategories)

	// Individual category operations
	dynamicCategory := categories.Group("/:id")
	dynamicCategory.PATCH("", h.UpdateCategory)
	dynamicCategory.DELETE("", h.DeleteCategory)
}