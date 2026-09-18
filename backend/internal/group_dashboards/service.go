package group_dashboards

import (
	"context"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// RawResourceRequest is a pending-resource-request count for one distinct faculty
// combination (a group's `faculty` column is a Postgres array, since a group can
// belong to more than one faculty), with array decoding already done at the
// repository boundary — the service only deals in decoded []string here.
type RawResourceRequest struct {
	Faculties []string
	Count     int
}

type Repository interface {
	GetGroupDashboardMetrics(ctx context.Context, groupID string) (upcoming int, pendingReports int, err error)
	GetFacultyDashboardMetrics(ctx context.Context, faculty string) (pendingRequests int, totalGroups int, err error)
	GetDeuDashboardMetrics(ctx context.Context) (pendingDeu int, rawResourceReqs []RawResourceRequest, activeGroups int, inactiveGroups int, err error)
}

type Service interface {
	GetGroupDashboard(ctx context.Context, groupID string) (*GroupDashboardResponse, error)
	GetFacultyDashboard(ctx context.Context, faculty string) (*FacultyDashboardResponse, error)
	GetDeuDashboard(ctx context.Context) (*DeuDashboardResponse, error)
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

func (s *service) GetGroupDashboard(ctx context.Context, groupID string) (*GroupDashboardResponse, error) {
	upcoming, pendingReports, err := s.repo.GetGroupDashboardMetrics(ctx, groupID)
	if err != nil {
		s.logger.Error("error retrieving group dashboard metrics", zap.String("group_id", groupID), zap.Error(err))
		return nil, err
	}

	return &GroupDashboardResponse{
		GroupID:            groupID,
		UpcomingActivities: upcoming,
		PendingReports:     pendingReports,
	}, nil
}

func (s *service) GetFacultyDashboard(ctx context.Context, faculty string) (*FacultyDashboardResponse, error) {
	pending, totalGroups, err := s.repo.GetFacultyDashboardMetrics(ctx, faculty)
	if err != nil {
		s.logger.Error("error retrieving faculty dashboard metrics", zap.String("faculty", faculty), zap.Error(err))
		return nil, err
	}

	return &FacultyDashboardResponse{
		Faculty:         faculty,
		PendingRequests: pending,
		TotalGroups:     totalGroups,
	}, nil
}

func (s *service) GetDeuDashboard(ctx context.Context) (*DeuDashboardResponse, error) {
	pendingDeu, rawReqs, activeGroups, inactiveGroups, err := s.repo.GetDeuDashboardMetrics(ctx)
	if err != nil {
		s.logger.Error("error retrieving DEU dashboard metrics", zap.Error(err))
		return nil, err
	}

	facultyMap := make(map[string]int)
	for _, item := range rawReqs {
		targetFaculty := resolveFaculty(item.Faculties)
		facultyMap[targetFaculty] += item.Count
	}

	processedReqs := make([]ResourceRequestsByFaculty, 0, len(facultyMap))
	for f, count := range facultyMap {
		processedReqs = append(processedReqs, ResourceRequestsByFaculty{
			Faculty: f,
			Count:   count,
		})
	}

	sort.Slice(processedReqs, func(i, j int) bool {
		return processedReqs[i].Count > processedReqs[j].Count
	})

	return &DeuDashboardResponse{
		PendingRequests:           pendingDeu,
		ResourceRequestsByFaculty: processedReqs,
		ActiveGroupsCount:         activeGroups,
		InactiveGroupsCount:       inactiveGroups,
	}, nil
}

// resolveFaculty buckets a group's faculty list into a single display faculty:
// groups with exactly one faculty are attributed to it, and groups with zero or
// multiple faculties (multidisciplinary) are bucketed under "DEU".
func resolveFaculty(faculties []string) string {
	var cleaned []string
	for _, f := range faculties {
		trimmed := strings.TrimSpace(f)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	if len(cleaned) != 1 {
		return "DEU"
	}

	return cleaned[0]
}
