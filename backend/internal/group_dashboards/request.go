package group_dashboards

type GroupDashboardRequest struct {
	GroupID string `param:"groupId" validate:"required,uuid"`
}

type FacultyDashboardRequest struct {
	Faculty string `param:"faculty" validate:"required"`
}
