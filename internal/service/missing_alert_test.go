package service_test

import (
	"errors"
	"path/filepath"
	"testing"

	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/service"
	"coldchain-alert/internal/store"
)

func TestMissingAlertSeverityReturnsNotFound(t *testing.T) {
	storage, err := store.Open(filepath.Join(t.TempDir(), "missing-alert.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	alerts := service.NewAlertService(storage)
	_, err = alerts.GetAlertSeverity("alert-does-not-exist")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing alert should be handleable: %v", err)
	}
}
