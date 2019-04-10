package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"os/exec"
	"path"
	"plugin"
	"strconv"

	"github.com/google/logger"

	"github.com/dergoegge/go-functions-sdk/internal/pkg/build"
	"github.com/dergoegge/go-functions-sdk/pkg/functions"
)

type gcloudProjectConfig struct {
	Configuration struct {
		Properties struct {
			Core struct {
				Project string `json:"project"`
			} `json:"core"`
		} `json:"properties"`
	} `json:"configuration"`
}

var projectID string

func setProjectID() error {
	cmd := exec.Command("gcloud", "config", "config-helper",
		"--format", "json")

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(outBuf.String())
	}

	var config gcloudProjectConfig
	err = json.Unmarshal(outBuf.Bytes(), &config)
	if err != nil {
		return nil
	}

	projectID = config.Configuration.Properties.Core.Project

	return nil
}

func firestoreDocToResource(doc string) string {
	return fmt.Sprintf("projects/%s/databases/(default)/documents/%s", projectID, doc)
}

func getDeployFlags(fnSym plugin.Symbol) ([]string, *functions.FunctionBuilder) {
	var triggerFlags []string
	var builder *functions.FunctionBuilder

	switch fnSym.(type) {
	case **functions.HTTPFunctionBuilder:
		httpBuilder := (*fnSym.(**functions.HTTPFunctionBuilder))
		builder = httpBuilder.FunctionBuilder
		triggerFlags = []string{"--trigger-http"}
		break

	case **functions.FirestoreFunctionBuilder:
		firestoreBuilder := (*fnSym.(**functions.FirestoreFunctionBuilder))
		builder = firestoreBuilder.FunctionBuilder
		triggerFlags = []string{
			"--trigger-event", firestoreBuilder.Event,
			"--trigger-resource", firestoreDocToResource(firestoreBuilder.Doc),
		}
		break
	case **functions.StorageFunctionBuilder:
		storageBuilder := (*fnSym.(**functions.StorageFunctionBuilder))
		builder = storageBuilder.FunctionBuilder
		triggerFlags = []string{
			"--trigger-event", storageBuilder.Event,
			"--trigger-resource", storageBuilder.GCBucket,
		}
		break
	}

	return triggerFlags, builder
}

func createDeployCommand(fnPlugin *plugin.Plugin, fn types.Object) (*exec.Cmd, error) {
	fnSym, err := fnPlugin.Lookup(fn.Name())
	if err != nil {
		return nil, err
	}

	deployFlags, fnBuilder := getDeployFlags(fnSym)

	cmdArgs := []string{
		"functions",
		"deploy",
		fn.Name(),
		"--runtime", "go111",
		"--source", "./" + fn.Pkg().Name(),
		"--entry-point", fn.Name() + ".Handler",
		"--region", fnBuilder.GCRegion,
		"--memory", fnBuilder.RuntimeOpts.Memory,
		"--timeout", strconv.Itoa(fnBuilder.RuntimeOpts.Timeout),
	}
	cmdArgs = append(cmdArgs, deployFlags...)

	return exec.Command("gcloud", cmdArgs...), nil
}

func Prepare(stagedFunctions []types.Object) ([]*exec.Cmd, error) {
	err := setProjectID()
	if err != nil {
		return nil, err
	}

	logger.Infof("Project: %s", projectID)

	cmds := []*exec.Cmd{}
	for _, fn := range stagedFunctions {
		pkg := fn.Pkg().Name()

		fnPlugin, err := plugin.Open(path.Join(build.PluginFolder, pkg+".so"))
		if err != nil {
			return nil, err
		}

		cmd, err := createDeployCommand(fnPlugin, fn)
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}