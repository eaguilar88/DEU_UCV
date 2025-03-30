package groups

import "github.com/eaguilar88/deu/pkg/entities"

// TODO: Implement request.go logic
type GetGroupRequest struct {
	ID string `param:"id" validate:"required"`
}
type GetGroupsRequest struct {
	PageScope entities.PageScope `query:"page_scope"`
}
type CreateGroupRequest struct {
	Group entities.ExtensionGroup `json:"group" validate:"required"`
}
type UpdateGroupRequest struct {
	ID    string                  `param:"id" validate:"required"`
	Group entities.ExtensionGroup `json:"group" validate:"required"`
}
type DeleteGroupRequest struct {
	ID string `param:"id" validate:"required"`
}
