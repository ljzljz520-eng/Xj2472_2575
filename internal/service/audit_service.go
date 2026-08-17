package service

import (
	"coldchain-alert/internal/domain"
	"coldchain-alert/internal/store"
)

type AuditService struct {
	storage *store.Store
}

func NewAuditService(storage *store.Store) *AuditService {
	return &AuditService{storage: storage}
}

func (s *AuditService) Record(actor, action, entity, entityID, detail, at string) (domain.AuditEntry, error) {
	entry := domain.AuditEntry{
		ID:        "audit-" + entity + "-" + entityID + "-" + action,
		Actor:     actor,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Detail:    detail,
		CreatedAt: at,
	}
	if err := s.storage.SaveAudit(entry); err != nil {
		return domain.AuditEntry{}, err
	}
	return entry, nil
}

func (s *AuditService) ForEntity(entity, entityID string) ([]domain.AuditEntry, error) {
	return s.storage.ListAuditsByEntity(entity, entityID)
}

func (s *AuditService) All() ([]domain.AuditEntry, error) {
	return s.storage.ListAudits()
}
