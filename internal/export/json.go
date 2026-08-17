package export

import (
	"encoding/json"
	"io"

	"coldchain-alert/internal/domain"
)

type Package struct {
	GeneratedFor string                  `json:"generated_for"`
	Alerts       []domain.ExportRow      `json:"alerts"`
	Summary      domain.DashboardSummary `json:"summary"`
}

func WriteJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func BuildPackage(summary domain.DashboardSummary, alerts []domain.ExportRow) Package {
	return Package{GeneratedFor: summary.Zone.ID, Alerts: alerts, Summary: summary}
}
