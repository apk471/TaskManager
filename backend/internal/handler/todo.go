package handler

import (
	"net/http"

	"github.com/apk471/go-taskmanager/internal/middleware"
	"github.com/apk471/go-taskmanager/internal/model"
	"github.com/apk471/go-taskmanager/internal/model/todo"
	"github.com/apk471/go-taskmanager/internal/server"
	"github.com/apk471/go-taskmanager/internal/service"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	Handler
	todoService *service.TodoService
}

func NewTodoHandler(s *server.Server, todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{
		Handler:     NewHandler(s),
		todoService: todoService,
	}
}

func (h *TodoHandler) CreateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.CreateTodo(c, userID, payload)
		},
		http.StatusCreated,
		&todo.CreateTodoPayload{},
	)(c)
}

func (h *TodoHandler) GetTodoByID(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetTodoByIDPayload) (*todo.PopulatedTodo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.GetTodoByID(c, userID, payload.ID)
		},
		http.StatusOK,
		&todo.GetTodoByIDPayload{},
	)(c)
}

func (h *TodoHandler) GetTodos(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, query *todo.GetTodosQuery) (*model.PaginatedResponse[todo.PopulatedTodo], error) {
			userID := middleware.GetUserID(c)
			return h.todoService.GetTodos(c, userID, query)
		},
		http.StatusOK,
		&todo.GetTodosQuery{},
	)(c)
}

func (h *TodoHandler) UpdateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoService.UpdateTodo(c, userID, payload)
		},
		http.StatusOK,
		&todo.UpdateTodoPayload{},
	)(c)
}

func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *todo.DeleteTodoPayload) error {
			userID := middleware.GetUserID(c)
			return h.todoService.DeleteTodo(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&todo.DeleteTodoPayload{},
	)(c)
}

// TODO: gettodostats route
