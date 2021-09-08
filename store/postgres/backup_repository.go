package postgres

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"

	"github.com/odpf/optimus/store"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/odpf/optimus/models"
	"github.com/pkg/errors"
)

type spec struct {
	result      interface{}
	description string
	config      models.DestinationConfig
}

type Backup struct {
	ID uuid.UUID `gorm:"primary_key;type:uuid"`

	ResourceID uuid.UUID
	Resource   Resource `gorm:"foreignKey:ResourceID"`

	Spec datatypes.JSON

	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

type backupRepository struct {
	db                      *gorm.DB
	namespace               models.NamespaceSpec
	datastore               models.Datastorer
	projectResourceSpecRepo store.ProjectResourceSpecRepository
}

func (r Backup) FromSpec(backupSpec models.BackupSpec) (Backup, error) {
	adaptResource, err := Resource{}.FromSpec(backupSpec.Resource)
	if err != nil {
		return Backup{}, err
	}

	toDBSpec := spec{
		result:      backupSpec.Result,
		description: backupSpec.Description,
		config:      backupSpec.Config,
	}
	specInBytes, err := json.Marshal(toDBSpec)
	if err != nil {
		return Backup{}, nil
	}

	return Backup{
		ID:         backupSpec.ID,
		ResourceID: adaptResource.ID,
		Resource:   adaptResource,
		Spec:       specInBytes,
	}, nil
}

func (repo *backupRepository) Insert(spec models.BackupSpec) error {
	if len(spec.Resource.ID) == 0 {
		return errors.New("resource cannot be empty")
	}
	p, err := Backup{}.FromSpec(spec)
	if err != nil {
		return err
	}
	return repo.db.Create(&p).Error
}

func NewBackupRepository(db *gorm.DB, namespace models.NamespaceSpec, ds models.Datastorer, projectResourceSpecRepo store.ProjectResourceSpecRepository) *resourceSpecRepository {
	return &resourceSpecRepository{
		db:                      db,
		namespace:               namespace,
		datastore:               ds,
		projectResourceSpecRepo: projectResourceSpecRepo,
	}
}
