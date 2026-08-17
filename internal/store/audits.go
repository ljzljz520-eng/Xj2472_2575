package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveAudit(entry domain.AuditEntry) error {
	if err := domain.RequireNonEmpty(entry.ID, entry.Actor, entry.Action, entry.Entity, entry.EntityID, entry.CreatedAt); err != nil {
		return err
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketAudits)).Get([]byte(entry.ID)) != nil {
			return fmt.Errorf("%w: audit %s", domain.ErrAlreadyExists, entry.ID)
		}
		return putJSON(tx, BucketAudits, entry.ID, entry)
	})
}

func (s *Store) ListAudits() ([]domain.AuditEntry, error) {
	var entries []domain.AuditEntry
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		entries, err = listJSON[domain.AuditEntry](tx, BucketAudits)
		return err
	})
	return entries, err
}

func (s *Store) ListAuditsByEntity(entity, entityID string) ([]domain.AuditEntry, error) {
	entries, err := s.ListAudits()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.AuditEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Entity == entity && entry.EntityID == entityID {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}
