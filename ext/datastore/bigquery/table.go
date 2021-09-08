package bigquery

import (
	"cloud.google.com/go/bigquery"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/googleapis/google-cloud-go-testing/bigquery/bqiface"
	"github.com/odpf/optimus/models"
	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

var (
	tableNameParseRegex = regexp.MustCompile(`^([\w-]+)\.(\w+)\.([\w-]+)$`)
)

func createTable(ctx context.Context, spec models.ResourceSpec, client bqiface.Client, upsert bool) error {
	bqResource, ok := spec.Spec.(BQTable)
	if !ok {
		return errors.New("failed to read table spec for bigquery")
	}

	// inherit from base
	bqResource.Metadata.Labels = spec.Labels

	dataset := client.DatasetInProject(bqResource.Project, bqResource.Dataset)
	if err := ensureDataset(ctx, dataset, BQDataset{
		Project:  bqResource.Project,
		Dataset:  bqResource.Dataset,
		Metadata: BQDatasetMetadata{},
	}, false); err != nil {
		return err
	}
	table := dataset.Table(bqResource.Table)
	return ensureTable(ctx, table, bqResource, upsert)
}

// ensureTable make sures table exists with provided config and update if required
func ensureTable(ctx context.Context, tableHandle bqiface.Table, t BQTable, upsert bool) error {
	meta, err := tableHandle.Metadata(ctx)
	if err != nil {
		if metaErr, ok := err.(*googleapi.Error); !ok || metaErr.Code != http.StatusNotFound {
			return err
		}
		m, err := bqCreateTableMetaAdapter(t)
		if err != nil {
			return err
		}
		return tableHandle.Create(ctx, m)
	}
	if !upsert {
		return nil
	}

	// update if already exists
	m, err := bqUpdateTableMetaAdapter(t)
	if err != nil {
		return err
	}
	_, err = tableHandle.Update(ctx, m, meta.ETag)
	return err
}

// getTable retrieves bq table information
func getTable(ctx context.Context, resourceSpec models.ResourceSpec, client bqiface.Client) (models.ResourceSpec, error) {
	var bqResource BQTable
	bqResource, ok := resourceSpec.Spec.(BQTable)
	if !ok {
		return models.ResourceSpec{}, errors.New("failed to read table spec for bigquery")
	}

	dataset := client.DatasetInProject(bqResource.Project, bqResource.Dataset)
	if _, err := dataset.Metadata(ctx); err != nil {
		return models.ResourceSpec{}, err
	}

	table := dataset.Table(bqResource.Table)
	tableMeta, err := table.Metadata(ctx)
	if err != nil {
		return models.ResourceSpec{}, err
	}

	// generate schema
	tableSchema, err := bqSchemaFrom(tableMeta.Schema)
	if err != nil {
		return models.ResourceSpec{}, err
	}

	// update metadata
	bqResource.Metadata = BQTableMetadata{
		Description: tableMeta.Description,
		Labels:      tableMeta.Labels,
		Schema:      tableSchema,
		Cluster:     bqClusteringFrom(tableMeta.Clustering),
		ViewQuery:   tableMeta.ViewQuery,
		Location:    tableMeta.Location,
	}

	// if table is partitioned
	if tableMeta.TimePartitioning != nil {
		bqResource.Metadata.Partition = bqPartitioningFrom(tableMeta.TimePartitioning)
	} else if tableMeta.RangePartitioning != nil {
		bqResource.Metadata.Partition = &BQPartitionInfo{
			Field: tableMeta.RangePartitioning.Field,
			Range: bqPartitioningRangeFrom(tableMeta.RangePartitioning.Range),
		}
	}

	resourceSpec.Spec = bqResource
	return resourceSpec, nil
}

func deleteTable(ctx context.Context, resourceSpec models.ResourceSpec, client bqiface.Client) error {
	bqTable, ok := resourceSpec.Spec.(BQTable)
	if !ok {
		return errors.New("failed to read table spec for bigquery")
	}
	dataset := client.DatasetInProject(bqTable.Project, bqTable.Dataset)
	if _, err := dataset.Metadata(ctx); err != nil {
		return err
	}

	table := dataset.Table(bqTable.Table)
	return table.Delete(ctx)
}

func backupTable(ctx context.Context, spec models.ResourceSpec, client bqiface.Client) error {
	bqResourceSrc, ok := spec.Spec.(BQTable)
	if !ok {
		return errors.New("failed to read table spec for bigquery")
	}

	// inherit from base
	bqResourceSrc.Metadata.Labels = spec.Labels

	//will be refactored as input
	prefixTableName := "backup"
	ttlInDays := time.Duration(30)
	backupID := ""
	bqResourceDst := BQTable{
		Project: bqResourceSrc.Project,
		Dataset: "optimus_backup",
		Table:   fmt.Sprintf("%s_%s_%s_%s", prefixTableName, bqResourceSrc.Dataset, bqResourceSrc.Table, backupID),
	}

	// make sure dataset is present
	datasetDst := client.DatasetInProject(bqResourceDst.Project, bqResourceDst.Dataset)
	if err := ensureDataset(ctx, datasetDst, BQDataset{
		Project:  bqResourceSrc.Project,
		Dataset:  bqResourceSrc.Dataset,
		Metadata: BQDatasetMetadata{},
	}, false); err != nil {
		return err
	}

	datasetSrc := client.DatasetInProject(bqResourceSrc.Project, bqResourceSrc.Dataset)
	if _, err := datasetSrc.Metadata(ctx); err != nil {
		return err
	}

	// duplicate table
	tableSrc := datasetSrc.Table(bqResourceSrc.Table)
	tableDst := datasetDst.Table(bqResourceDst.Table)

	copier := tableDst.CopierFrom(tableSrc)
	job, err := copier.Run(ctx)
	if err != nil {
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return err
	}
	if err := status.Err(); err != nil {
		return err
	}

	// update expiry
	meta, err := tableDst.Metadata(ctx)
	if err != nil {
		return err
	}
	update := bigquery.TableMetadataToUpdate{
		ExpirationTime: time.Now().Add(time.Hour * 24 * ttlInDays),
	}
	if _, err = tableDst.Update(ctx, update, meta.ETag); err != nil {
		return err
	}

	return ensureTable(ctx, tableDst, bqResourceDst, false)
}
