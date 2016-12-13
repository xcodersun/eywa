package presenters

import (
	. "github.com/eywa/models"
)

type DashboardBrief struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewDashboardBrief(d *Dashboard) *DashboardBrief {
	return &DashboardBrief{
		ID:          d.Id,
		Name:        d.Name,
		Description: d.Description,
	}
}
