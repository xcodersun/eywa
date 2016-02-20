package presenters

import (
	"encoding/base64"
	. "github.com/vivowares/eywa/models"
	"strconv"
)

type DashboardBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DashboardDetail struct {
	ID string `json:"id"`
	*Dashboard
}

func NewDashboardBrief(d *Dashboard) *DashboardBrief {
	return &DashboardBrief{
		ID:          base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(d.Id))),
		Name:        d.Name,
		Description: d.Description,
	}
}

func NewDashboardDetail(d *Dashboard) *DashboardDetail {
	return &DashboardDetail{
		ID:        base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(d.Id))),
		Dashboard: d,
	}
}
