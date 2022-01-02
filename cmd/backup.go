package cmd

import (
	"time"

	"github.com/odpf/optimus/config"
	"github.com/odpf/optimus/models"
	"github.com/odpf/salt/log"
	cli "github.com/spf13/cobra"
)

var (
	backupTimeout = time.Minute * 15
)

func backupCommand(l log.Logger, datastoreRepo models.DatastoreRepo, conf config.Provider) *cli.Command {
	cmd := &cli.Command{
		Use:   "backup",
		Short: "Backup a resource and its downstream",
		Long: `Backup supported resource of a datastore and all of its downstream dependencies.
Operation can take upto few minutes to complete. It is advised to check the operation status
using "list" command.
`,
	}
	cmd.AddCommand(backupResourceSubCommand(l, datastoreRepo, conf))
	cmd.AddCommand(backupListSubCommand(l, datastoreRepo, conf))
	return cmd
}
