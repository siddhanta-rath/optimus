package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/AlecAivazis/survey/v2"
	"github.com/odpf/optimus/config"
	"github.com/odpf/optimus/models"
	"github.com/odpf/salt/log"
	cli "github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	defaultHost = "localhost"
)

func configCommand(l log.Logger, dsRepo models.DatastoreRepo) *cli.Command {
	c := &cli.Command{
		Use:   "config",
		Short: "Manage optimus configuration required to deploy specifications",
	}
	c.AddCommand(configInitCommand(l, dsRepo))
	return c
}

func configInitCommand(l log.Logger, dsRepo models.DatastoreRepo) *cli.Command {
	c := &cli.Command{
		Use:   "init",
		Short: "Initialize optimus configuration file",
		RunE: func(c *cli.Command, args []string) (err error) {
			conf := config.Optimus{
				Version: 1,
				Host:    defaultHost,
			}
			questions := []*survey.Question{
				{
					Name: "ProjectName",
					Prompt: &survey.Input{
						Message: "What is the project name?",
						Help:    "project name of the repository",
					},
					Validate: survey.Required,
				},
				{
					Name: "RegisterProjectConfig",
					Prompt: &survey.Select{
						Message: "Register project configs?",
						Options: []string{"Yes", "No"},
						Default: "No",
					},
				},
			}
			answers := map[string]interface{}{}
			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}
			conf.Project.Name = answers["ProjectName"].(string)

			// for project config
			if option, ok := answers["RegisterProjectConfig"]; ok && option.(survey.OptionAnswer).Value == "Yes" {
				conf, err = projectConfigQuestions(conf)
				if err != nil {
					return err
				}
			}

			// questions for namespace config
			questions = []*survey.Question{
				{
					Name: "NamespaceName",
					Prompt: &survey.Input{
						Message: "What is the namespace name?",
						Help:    "specific namespace name for jobs and resources inside the directory",
					},
					Validate: survey.Required,
				},
				{
					Name: "RegisterNamespaceConfig",
					Prompt: &survey.Select{
						Message: "Register namespace configs?",
						Options: []string{"Yes", "No"},
						Default: "No",
					},
				},
			}
			answers = map[string]interface{}{}
			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}

			conf.Namespace.Name = answers["NamespaceName"].(string)
			// for namespace config
			if option, ok := answers["RegisterNamespaceConfig"]; ok && option.(survey.OptionAnswer).Value == "Yes" {
				conf, err = namespaceConfigQuestions(conf)
				if err != nil {
					return err
				}
			}

			// for datastore
			questions = []*survey.Question{
				{
					Name: "JobPath",
					Prompt: &survey.Input{
						Message: "Scheduled jobs directory",
						Default: "./jobs",
						Help:    "relative directory path to jobs specification",
					},
					Validate: survey.Required,
				},
				{
					Name: "RegisterDatastore",
					Prompt: &survey.Select{
						Message: "Register datastore configs?",
						Options: []string{"Yes", "No"},
						Default: "No",
					},
				},
			}
			answers = map[string]interface{}{}
			if err := survey.Ask(questions, &answers); err != nil {
				return err
			}
			conf.Namespace.Job.Path = answers["JobPath"].(string)
			if option, ok := answers["RegisterDatastore"]; ok && option.(survey.OptionAnswer).Value == "Yes" {
				conf, err = datastoreConfigQuestions(conf, dsRepo)
				if err != nil {
					return err
				}
			}

			confMarshaled, err := yaml.Marshal(conf)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(fmt.Sprintf("%s.%s", config.FileName, config.FileExtension), confMarshaled, 0655); err != nil {
				return err
			}
			l.Info("configuration initialised successfully")

			return nil
		},
	}
	return c
}

func projectConfigQuestions(conf config.Optimus) (config.Optimus, error) {
	conf.Project.Config = map[string]string{}
	registerMore := "Yes"
	for registerMore == "Yes" {
		configAnswers := map[string]interface{}{}
		if err := survey.Ask([]*survey.Question{
			{
				Name: "Name",
				Prompt: &survey.Input{
					Message: "Name of the config",
				},
				Validate: survey.MinLength(3),
			},
			{
				Name: "Value",
				Prompt: &survey.Input{
					Message: "Value",
				},
				Validate: survey.MinLength(1),
			},
		}, &configAnswers); err != nil {
			return conf, err
		}

		if err := survey.AskOne(&survey.Select{
			Message: "Add one more?",
			Options: []string{"Yes", "No"},
			Default: "Yes",
		}, &registerMore); err != nil {
			return conf, err
		}
		conf.Project.Config[configAnswers["Name"].(string)] = configAnswers["Value"].(string)
	}

	return conf, nil
}

func namespaceConfigQuestions(conf config.Optimus) (config.Optimus, error) {
	conf.Namespace.Config = map[string]string{}
	registerMore := "Yes"
	for registerMore == "Yes" {
		configAnswers := map[string]interface{}{}
		if err := survey.Ask([]*survey.Question{
			{
				Name: "Name",
				Prompt: &survey.Input{
					Message: "Name of the config",
				},
				Validate: survey.MinLength(3),
			},
			{
				Name: "Value",
				Prompt: &survey.Input{
					Message: "Value",
				},
				Validate: survey.MinLength(1),
			},
		}, &configAnswers); err != nil {
			return conf, err
		}

		if err := survey.AskOne(&survey.Select{
			Message: "Add one more?",
			Options: []string{"Yes", "No"},
			Default: "Yes",
		}, &registerMore); err != nil {
			return conf, err
		}
		conf.Namespace.Config[configAnswers["Name"].(string)] = configAnswers["Value"].(string)
	}

	return conf, nil
}

func datastoreConfigQuestions(conf config.Optimus, dsRepo models.DatastoreRepo) (config.Optimus, error) {
	dsOptions := []string{}
	for _, ds := range dsRepo.GetAll() {
		dsOptions = append(dsOptions, ds.Name())
	}
	conf.Namespace.Datastore = []config.Datastore{}

	configAnswers := map[string]interface{}{}
	if err := survey.Ask([]*survey.Question{
		{
			Name: "Type",
			Prompt: &survey.Select{
				Message: "Type of the datastore",
				Options: dsOptions,
			},
		},
		{
			Name: "Path",
			Prompt: &survey.Input{
				Message: "Path for specifications",
			},
			Validate: survey.MinLength(1),
		},
	}, &configAnswers); err != nil {
		return conf, err
	}
	conf.Namespace.Datastore = append(conf.Namespace.Datastore, config.Datastore{
		Type: configAnswers["Type"].(survey.OptionAnswer).Value,
		Path: configAnswers["Path"].(string),
	})

	return conf, nil
}
