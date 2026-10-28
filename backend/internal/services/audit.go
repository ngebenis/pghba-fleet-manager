package services

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pghba-manager/backend/internal/models"
	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(
	userID *uuid.UUID,
	username string,
	event models.AuditEvent,
	target string,
	instanceID *uuid.UUID,
	instanceLabel string,
	clientIP string,
	detail interface{},
) {
	// Kolom detail bertipe jsonb — harus valid JSON, tidak boleh string kosong
	detailJSON := "null"
	if detail != nil {
		b, err := json.Marshal(detail)
		if err == nil {
			detailJSON = string(b)
		}
	}

	log := &models.AuditLog{
		UserID:        userID,
		Username:      username,
		Event:         event,
		Target:        target,
		InstanceID:    instanceID,
		InstanceLabel: instanceLabel,
		ClientIP:      clientIP,
		Detail:        detailJSON,
	}
	s.db.Create(log)
}
