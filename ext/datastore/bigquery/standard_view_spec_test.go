package bigquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardViewSpecHandler(t *testing.T) {
	t.Run("should generate urn successfully", func(t *testing.T) {
		project := "sample-project"
		dataset := "sample-dataset"
		table := "sample-table"

		urn, err := standardViewSpec{}.GenerateURN(BQTable{
			Project: project,
			Dataset: dataset,
			Table:   table,
		})

		assert.Nil(t, err)
		assert.Equal(t, "bigquery://sample-project:sample-dataset.sample-table", urn)
	})
}
