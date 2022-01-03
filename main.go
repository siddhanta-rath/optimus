package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	_ "net/http/pprof"

	"github.com/hashicorp/go-hclog"
	hPlugin "github.com/hashicorp/go-plugin"
	"github.com/odpf/optimus/cmd"
	"github.com/odpf/optimus/config"
	_ "github.com/odpf/optimus/ext/datastore"
	"github.com/odpf/optimus/models"
	"github.com/odpf/optimus/plugin"
	_ "github.com/odpf/optimus/plugin"
	"github.com/odpf/salt/log"
)

var (
	errRequestFail = errors.New("🔥 unable to complete request successfully")
)

type PlainFormatter struct{}

func (p *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if len(entry.Data) > 0 {
		var data string
		for key, val := range entry.Data {
			data += fmt.Sprintf("%s: %v ", key, val)
		}
		return []byte(fmt.Sprintf("%s %s\n", entry.Message, data)), nil
	}
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	configuration, err := config.InitOptimus()
	if err != nil {
		fmt.Printf("ERROR: %s", err.Error())
		os.Exit(1)
	}

	var jsonLogger log.Logger
	var plainLogger log.Logger
	pluginLogLevel := hclog.Info
	if configuration.GetLog().Level != "" {
		jsonLogger = log.NewLogrus(log.LogrusWithLevel(configuration.GetLog().Level), log.LogrusWithWriter(os.Stderr))
		plainLogger = log.NewLogrus(log.LogrusWithLevel(configuration.GetLog().Level), log.LogrusWithFormatter(new(PlainFormatter)))
		if strings.ToLower(configuration.GetLog().Level) == "debug" {
			pluginLogLevel = hclog.Debug
		}
	} else {
		jsonLogger = log.NewLogrus(log.LogrusWithLevel("INFO"), log.LogrusWithWriter(os.Stderr))
		plainLogger = log.NewLogrus(log.LogrusWithLevel("INFO"), log.LogrusWithFormatter(new(PlainFormatter)))
	}

	// init telemetry
	teleShutdown, err := config.InitTelemetry(jsonLogger, configuration)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	defer teleShutdown()

	// discover and load plugins
	if err := plugin.Initialize(hclog.New(&hclog.LoggerOptions{
		Name:   "optimus",
		Output: os.Stdout,
		Level:  pluginLogLevel,
	})); err != nil {
		hPlugin.CleanupClients()
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	// Make sure we clean up any managed plugins at the end of this
	defer hPlugin.CleanupClients()

	command := cmd.New(
		plainLogger,
		jsonLogger,
		configuration,
		models.PluginRegistry,
		models.DatastoreRegistry,
	)
	if err := command.Execute(); err != nil {
		hPlugin.CleanupClients()
		// no need to print err here, `command` does that already
		fmt.Println(errRequestFail)
		os.Exit(1)
	}
}
