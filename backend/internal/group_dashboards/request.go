package group_dashboards

type GroupDashboardRequest struct {
	GroupID string `param:"groupId" validate:"required,numeric"`
}

type FacultyDashboardRequest struct {
	Faculty string `param:"faculty" validate:"required"`
}
