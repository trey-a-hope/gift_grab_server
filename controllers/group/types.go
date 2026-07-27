// Package group contains controllers and types for querying and managing Nakama groups.
package group

import "github.com/heroiclabs/nakama-common/api"

// GroupRequest represents the standard request payload for group operations containing a group ID.
type GroupRequest struct {
	GroupId string `json:"group_id"`
}

// GroupResponse defines a clean, JSON-serializable representation of a Nakama group object.
type GroupResponse struct {
	ID          string `json:"id"`
	CreatorID   string `json:"creator_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	LangTag     string `json:"lang_tag"`
	Metadata    string `json:"metadata"`
	AvatarURL   string `json:"avatar_url"`
	Open        bool   `json:"open"`
	EdgeCount   int32  `json:"edge_count"`
	MaxCount    int32  `json:"max_count"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}

// NewGroupResponse converts an api.Group object from Nakama's library to a formatted GroupResponse.
func NewGroupResponse(group *api.Group) GroupResponse {
	return GroupResponse{
		ID:          group.Id,
		CreatorID:   group.CreatorId,
		Name:        group.Name,
		Description: group.Description,
		LangTag:     group.LangTag,
		Metadata:    group.Metadata,
		AvatarURL:   group.AvatarUrl,
		Open:        group.Open.GetValue(),
		EdgeCount:   group.EdgeCount,
		MaxCount:    group.MaxCount,
		CreateTime:  group.CreateTime.AsTime().Format("2006-01-02T15:04:05Z"),
		UpdateTime:  group.UpdateTime.AsTime().Format("2006-01-02T15:04:05Z"),
	}
}
