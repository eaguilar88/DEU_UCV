package group_requests

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	ApproveGroupRequest(ctx context.Context, reqID string) error
	RejectGroupRequest(ctx context.Context, reqID string) error
	GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error)
	GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error)
	GetGroupRequestsByGroupID(ctx context.Context, groupID string) ([]entities.GroupRequest, error)
	GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error)
	ActivateGroup(ctx context.Context, groupID string) error
	CreateUser(ctx context.Context, user entities.User) (int64, error)
	AddRoleToUser(ctx context.Context, tx *sql.Tx, userID string, role int) error
	UpdateGroupUserID(ctx context.Context, groupID string, userID string) error
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) ApproveGroupRequest(ctx context.Context, reqID string) error {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return err
	}

	if err := s.repo.ApproveGroupRequest(ctx, reqID); err != nil {
		return err
	}

	allRequests, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to fetch group requests for activation check", zap.Error(err), zap.String("group_id", req.GroupID))
		return err
	}

	allApproved := true
	for _, r := range allRequests {
		if r.ID == reqID {
			continue
		}
		if string(r.Status) != "approved" {
			allApproved = false
			break
		}
	}

	if allApproved {
		s.logger.Info("all requests approved for group, creating group admin user and activating group", zap.String("group_id", req.GroupID))

		// Generación de contraseña temporal e impresión en consola para pruebas
		rawPassword := fmt.Sprintf("AdminGroup_%s_2026!", req.GroupID)
		fmt.Printf("[DEBUG] Contraseña generada para el admin del grupo %s: %s\n", req.GroupID, rawPassword)

		hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error("failed to hash password for new group admin", zap.Error(err))
			return err
		}
		hashedPassword := string(hashedPasswordBytes)

		groupIDInt, _ := strconv.Atoi(req.GroupID)
		cleanGroupName := strings.ToLower(strings.ReplaceAll(req.GroupName, " ", "_"))

		adminUser := entities.User{
			CI:             fmt.Sprintf("%d", 99000000+groupIDInt),
			Email:          fmt.Sprintf("%s@extension.ucv.ve", cleanGroupName),
			FirstName:      "Representante",
			LastName:       req.GroupName,
			Password:       hashedPassword,
			DateOfBirth:    "2000-01-01",
    		EducationLevel: "bachiller",
			Roles: []string{
				entities.RoleNameFromID(entities.RoleExtension),
			},
		}

		newUserID, err := s.repo.CreateUser(ctx, adminUser)
		if err != nil {
			s.logger.Error("failed to create user for group admin", zap.Error(err))
			return err
		}

		userIDStr := fmt.Sprintf("%d", newUserID)

		// Se asigna el rol de group_admin usando AddRoleToUser
		/*err = s.repo.AddRoleToUser(ctx, nil, userIDStr, entities.RoleAdminGrupo)
		if err != nil {
			s.logger.Error("failed to assign extension role to user", zap.Error(err))
			return err
		}*/

		// Actualizar el user_id asignado en deu.extension_groups
		if err := s.repo.UpdateGroupUserID(ctx, req.GroupID, userIDStr); err != nil {
			s.logger.Error("failed to update group user_id", zap.Error(err), zap.String("group_id", req.GroupID))
			return err
		}

		if err := s.repo.ActivateGroup(ctx, req.GroupID); err != nil {
			s.logger.Error("failed to activate group", zap.Error(err), zap.String("group_id", req.GroupID))
			return err
		}
	}

	return nil
}

func (s *service) RejectGroupRequest(ctx context.Context, reqID string) error {
	return s.repo.RejectGroupRequest(ctx, reqID)
}

func (s *service) GetGroupRequestsByFaculty(ctx context.Context, faculty entities.Faculty, status string, pageScope entities.PageScope) ([]entities.GroupRequest, entities.PageScope, int, error) {
	requests, ps, pendingCount, err := s.repo.GetGroupRequestsByFaculty(ctx, faculty, status, pageScope)
	if err != nil {
		s.logger.Error("failed to get group requests by faculty", zap.Error(err))
		return nil, entities.PageScope{}, 0, err
	}

	for i, req := range requests {
		approvals, err := s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
		if err != nil {
			s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
			return nil, entities.PageScope{}, 0, err
		}
		requests[i].Approvals = approvals
	}

	return requests, ps, pendingCount, nil
}

func (s *service) GetGroupRequestByID(ctx context.Context, reqID string) (entities.GroupRequest, error) {
	req, err := s.repo.GetGroupRequestByID(ctx, reqID)
	if err != nil {
		return entities.GroupRequest{}, err
	}

	req.Approvals, err = s.repo.GetGroupRequestsByGroupID(ctx, req.GroupID)
	if err != nil {
		s.logger.Error("failed to get approvals for group request", zap.Error(err), zap.String("group_id", req.GroupID))
		return entities.GroupRequest{}, err
	}

	return req, nil
}

func (s *service) GetPendingGroupRequestsCounts(ctx context.Context, faculty entities.Faculty) ([]entities.FacultyPendingCount, error) {
	return s.repo.GetPendingGroupRequestsCounts(ctx, faculty)
}
