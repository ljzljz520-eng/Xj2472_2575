package export

import (
	"encoding/csv"
	"io"
	"strconv"

	"coldchain-alert/internal/domain"
)

func AlertRows(alerts []domain.TemperatureAlert, zones map[string]domain.Zone) []domain.ExportRow {
	rows := make([]domain.ExportRow, 0, len(alerts))
	for _, alert := range alerts {
		zoneName := alert.ZoneID
		if zone, ok := zones[alert.ZoneID]; ok {
			zoneName = zone.Name
		}
		rows = append(rows, domain.ExportRow{AlertID: alert.ID, Zone: zoneName, Severity: string(alert.Severity), Status: string(alert.Status), Reason: alert.Reason, OpenedAt: alert.OpenedAt, UpdatedAt: alert.UpdatedAt})
	}
	return rows
}

func WriteAlertCSV(writer io.Writer, rows []domain.ExportRow) error {
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"alert_id", "zone", "severity", "status", "reason", "opened_at", "updated_at"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := output.Write([]string{row.AlertID, row.Zone, row.Severity, row.Status, row.Reason, row.OpenedAt, row.UpdatedAt}); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}

func WriteDashboardCSV(writer io.Writer, summary domain.DashboardSummary) error {
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"zone", "temperature", "humidity", "open_alerts", "critical_alerts", "door_open_count", "online_devices", "total_devices", "health_score"}); err != nil {
		return err
	}
	temperature := ""
	humidity := ""
	if summary.LatestReading != nil {
		temperature = strconv.FormatFloat(summary.LatestReading.Temperature, 'f', 1, 64)
		humidity = strconv.FormatFloat(summary.LatestReading.Humidity, 'f', 1, 64)
	}
	if err := output.Write([]string{summary.Zone.Name, temperature, humidity, strconv.Itoa(summary.OpenAlertCount), strconv.Itoa(summary.CriticalCount), strconv.Itoa(summary.DoorOpenCount), strconv.Itoa(summary.OnlineDevices), strconv.Itoa(summary.TotalDevices), strconv.Itoa(summary.HealthScore)}); err != nil {
		return err
	}
	output.Flush()
	return output.Error()
}
