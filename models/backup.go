package models

import (
	"github.com/google/uuid"
)

type BackupRequest struct {
	ResourceName     string
	Project          ProjectSpec
	Datastore        string
	Description      string
	IgnoreDownstream bool
}

type DestinationConfig struct {
	TTLInDays   int
	Dataset     string
	TablePrefix string
}

type BackupSpec struct {
	ID          uuid.UUID
	Resource    ResourceSpec
	Result      interface{}
	Description string
	Config      DestinationConfig
}
