package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bekbull/online-shop/services/user/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// HTTPServer handles HTTP requests for the User service
type HTTPServer struct {
	router      *chi.Mux
	userService domain.UserService
}

// NewHTTPServer creates a new HTTP server for the User service
func NewHTTPServer(userService domain.UserService) *HTTPServer {
	server := &HTTPServer{
		router:      chi.NewRouter(),
		userService: userService,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures the HTTP routes
func (s *HTTPServer) setupRoutes() {
	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)

	// API Routes with versioning
	s.router.Route("/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.ListUsers)
			r.Post("/", s.CreateUser)
			r.Get("/{id}", s.GetUser)
			r.Put("/{id}", s.UpdateUser)
			r.Delete("/{id}", s.DeleteUser)
		})
	})

	// Health check endpoint
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// Router returns the HTTP router
func (s *HTTPServer) Router() http.Handler {
	return s.router
}

// CreateUser handles user creation requests
func (s *HTTPServer) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string   `json:"email"`
		FirstName string   `json:"first_name"`
		LastName  string   `json:"last_name"`
		Password  string   `json:"password"`
		Roles     []string `json:"roles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.userService.CreateUser(req.Email, req.FirstName, req.LastName, req.Password, req.Roles)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondWithJSON(w, http.StatusCreated, mapUserToResponse(user))
}

// GetUser handles user retrieval requests
func (s *HTTPServer) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := s.userService.GetUser(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUserToResponse(user))
}

// UpdateUser handles user update requests
func (s *HTTPServer) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Email     *string  `json:"email,omitempty"`
		FirstName *string  `json:"first_name,omitempty"`
		LastName  *string  `json:"last_name,omitempty"`
		Password  *string  `json:"password,omitempty"`
		Roles     []string `json:"roles,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.FirstName != nil {
		updates["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updates["last_name"] = *req.LastName
	}
	if req.Password != nil {
		updates["password"] = *req.Password
	}
	if req.Roles != nil {
		updates["roles"] = req.Roles
	}

	user, err := s.userService.UpdateUser(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondWithJSON(w, http.StatusOK, mapUserToResponse(user))
}

// DeleteUser handles user deletion requests
func (s *HTTPServer) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.userService.DeleteUser(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUsers handles requests to list users
func (s *HTTPServer) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	page := 1
	pageSize := 10
	emailFilter := r.URL.Query().Get("email")

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	users, total, err := s.userService.ListUsers(page, pageSize, emailFilter)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	responseUsers := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		responseUsers = append(responseUsers, mapUserToResponse(user))
	}

	response := map[string]interface{}{
		"users":       responseUsers,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// respondWithJSON writes a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// mapUserToResponse maps a domain User to a response object
func mapUserToResponse(user *domain.User) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"roles":      user.Roles,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}
