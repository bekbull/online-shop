package handler

import (
	"context"
	"time"

	pb "github.com/bekbull/online-shop/services/user/api/proto"
	"github.com/bekbull/online-shop/services/user/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer is the gRPC server for the User service
type GRPCServer struct {
	pb.UnimplementedUserServiceServer
	userService domain.UserService
}

// NewGRPCServer creates a new gRPC server for the User service
func NewGRPCServer(userService domain.UserService) *GRPCServer {
	return &GRPCServer{
		userService: userService,
	}
}

// CreateUser creates a new user
func (s *GRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.CreateUser(
		req.Email,
		req.FirstName,
		req.LastName,
		req.Password,
		req.Roles,
	)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to create user: %v", err)
	}

	return convertDomainUserToProto(user), nil
}

// GetUser retrieves a user by ID
func (s *GRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.GetUser(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get user: %v", err)
	}

	return convertDomainUserToProto(user), nil
}

// UpdateUser updates an existing user
func (s *GRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
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
	if len(req.Roles) > 0 {
		updates["roles"] = req.Roles
	}

	user, err := s.userService.UpdateUser(req.Id, updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return convertDomainUserToProto(user), nil
}

// DeleteUser deletes a user by ID
func (s *GRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(req.Id)
	if err != nil {
		return &pb.DeleteUserResponse{Success: false}, status.Errorf(codes.NotFound, "failed to delete user: %v", err)
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// ListUsers retrieves a list of users with pagination and optional filtering
func (s *GRPCServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, total, err := s.userService.ListUsers(int(req.Page), int(req.PageSize), req.EmailFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var protoUsers []*pb.UserResponse
	for _, user := range users {
		protoUsers = append(protoUsers, convertDomainUserToProto(user))
	}

	return &pb.ListUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetUserByEmail retrieves a user by email
func (s *GRPCServer) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.UserResponse, error) {
	user, err := s.userService.GetUserByEmail(req.Email)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get user by email: %v", err)
	}

	return convertDomainUserToProto(user), nil
}

// convertDomainUserToProto converts a domain User to a proto UserResponse
func convertDomainUserToProto(user *domain.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
