package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/odpf/optimus/models"
	"github.com/odpf/optimus/store"
	"github.com/odpf/optimus/store/local"
	"github.com/odpf/optimus/utils"
	"github.com/odpf/salt/log"
	"github.com/spf13/afero"
	cli "github.com/spf13/cobra"
)

var (
	validateResourceName = utils.ValidatorFactory.NewFromRegex(`^[a-zA-Z0-9][a-zA-Z0-9_\-\.]+$`,
		`invalid name (can only contain characters A-Z (in either case), 0-9, "-", "_" or "." and must start with an alphanumeric character)`)
)

func resourceCommand(l log.Logger, datastoreSpecsFs map[string]afero.Fs, datastoreRepo models.DatastoreRepo) *cli.Command {
	cmd := &cli.Command{
		Use:   "resource",
		Short: "Interact with data resource",
	}
	cmd.AddCommand(createResourceSubCommand(l, datastoreSpecsFs, datastoreRepo))
	return cmd
}

func createResourceSubCommand(l log.Logger, datastoreSpecFs map[string]afero.Fs, datastoreRepo models.DatastoreRepo) *cli.Command {
	return &cli.Command{
		Use:   "resource",
		Short: "Create a new resource",
		RunE: func(cmd *cli.Command, args []string) error {
			availableStorer := []string{}
			for _, s := range datastoreRepo.GetAll() {
				availableStorer = append(availableStorer, s.Name())
			}
			var storerName string
			if err := survey.AskOne(&survey.Select{
				Message: "Select supported datastores?",
				Options: availableStorer,
			}, &storerName); err != nil {
				return err
			}
			repoFS, ok := datastoreSpecFs[storerName]
			if !ok {
				return fmt.Errorf("unregistered datastore, please use configuration file to set datastore path")
			}

			// find requested datastore
			availableTypes := []string{}
			datastore, _ := datastoreRepo.GetByName(storerName)
			for dsType := range datastore.Types() {
				availableTypes = append(availableTypes, dsType.String())
			}
			resourceSpecRepo := local.NewResourceSpecRepository(repoFS, datastore)

			// find resource type
			var resourceType string
			if err := survey.AskOne(&survey.Select{
				Message: "Select supported resource type?",
				Options: availableTypes,
			}, &resourceType); err != nil {
				return err
			}
			typeController, _ := datastore.Types()[models.ResourceType(resourceType)]

			// find directory to store spec
			rwd, err := getWorkingDirectory(repoFS, "")
			if err != nil {
				return err
			}
			newDirName, err := getDirectoryName(rwd)
			if err != nil {
				return err
			}

			resourceDirectory := filepath.Join(rwd, newDirName)
			resourceNameDefault := strings.ReplaceAll(strings.ReplaceAll(resourceDirectory, "/", "."), "\\", ".")

			var qs = []*survey.Question{
				{
					Name: "name",
					Prompt: &survey.Input{
						Message: "What is the resource name?(should conform to selected resource type)",
						Default: resourceNameDefault,
					},
					Validate: survey.ComposeValidators(validateNoSlash, survey.MinLength(3),
						survey.MaxLength(1024), IsValidDatastoreSpec(typeController.Validator()),
						IsResourceNameUnique(resourceSpecRepo)),
				},
			}
			inputs := map[string]interface{}{}
			if err := survey.Ask(qs, &inputs); err != nil {
				return err
			}
			resourceName := inputs["name"].(string)

			if err := resourceSpecRepo.SaveAt(models.ResourceSpec{
				Version:   1,
				Name:      resourceName,
				Type:      models.ResourceType(resourceType),
				Datastore: datastore,
				Assets:    typeController.DefaultAssets(),
			}, resourceDirectory); err != nil {
				return err
			}

			l.Info(fmt.Sprintf("resource created successfully %s", resourceName))
			return nil
		},
	}
}

// IsResourceNameUnique return a validator that checks if the resource already exists with the same name
func IsResourceNameUnique(repository store.ResourceSpecRepository) survey.Validator {
	return func(val interface{}) error {
		if str, ok := val.(string); ok {
			if _, err := repository.GetByName(context.Background(), str); err == nil {
				return fmt.Errorf("resource with the provided name already exists")
			} else if err != models.ErrNoSuchSpec && err != models.ErrNoResources {
				return err
			}
		} else {
			// otherwise we cannot convert the value into a string and cannot find a resource name
			return fmt.Errorf("invalid type of resource name %v", reflect.TypeOf(val).Name())
		}
		// the input is fine
		return nil
	}
}

// IsValidDatastoreSpec tries to adapt provided resource with datastore
func IsValidDatastoreSpec(valiFn models.DatastoreSpecValidator) survey.Validator {
	return func(val interface{}) error {
		if str, ok := val.(string); ok {
			if err := valiFn(models.ResourceSpec{
				Name: str,
			}); err != nil {
				return err
			}
		} else {
			// otherwise we cannot convert the value into a string and cannot find a resource name
			return fmt.Errorf("invalid type of resource name %v", reflect.TypeOf(val).Name())
		}
		// the input is fine
		return nil
	}
}
