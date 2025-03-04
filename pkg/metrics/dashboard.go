package metrics

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
)

type Dashboard struct {
	bun.BaseModel `bun:"dashboards,alias:d"`

	ID         uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID  uint32 `json:"projectId"`
	TemplateID string `json:"templateId" bun:",nullzero"`

	Name      string `json:"name"`
	BaseQuery string `json:"baseQuery" bun:",nullzero"`

	IsTable bool                     `json:"isTable" bun:",nullzero"`
	Metrics []upql.Metric            `json:"metrics" bun:",nullzero"`
	Query   string                   `json:"query" bun:",nullzero"`
	Columns map[string]*MetricColumn `json:"columnMap" bun:",nullzero"`
}

func (d *Dashboard) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}
	return nil
}

func SelectDashboard(ctx context.Context, app *bunapp.App, id uint64) (*Dashboard, error) {
	dash := new(Dashboard)
	if err := app.DB.NewSelect().
		Model(dash).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return dash, nil
}

func SelectDashboardMap(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (map[string]*Dashboard, error) {
	var dashboards []*Dashboard

	if err := app.DB.NewSelect().
		Model(&dashboards).
		Where("project_id = ?", projectID).
		Where("template_id IS NOT NULL").
		Scan(ctx); err != nil {
		return nil, err
	}

	m := make(map[string]*Dashboard, len(dashboards))

	for _, dash := range dashboards {
		m[dash.TemplateID] = dash
	}

	return m, nil
}

func InsertDashboard(ctx context.Context, app *bunapp.App, dash *Dashboard) error {
	if dash.Columns == nil {
		dash.Columns = make(map[string]*MetricColumn)
	}

	if _, err := app.DB.NewInsert().
		Model(dash).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func DeleteDashboard(ctx context.Context, app *bunapp.App, id uint64) error {
	if _, err := app.DB.NewDelete().
		Model((*Dashboard)(nil)).
		Where("id = ?", id).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
