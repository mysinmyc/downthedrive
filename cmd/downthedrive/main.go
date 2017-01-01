package main

import (
	"fmt"
	"os"
	"io"
	"github.com/mysinmyc/downthedrive/cmdactions"
	"github.com/mysinmyc/gocommons/diagnostic"
)

func usage() {
	fmt.Printf("USAGE: %s {action} [arguments ...]\n\nActions available:\n", os.Args[0])

	for vCurActionName, vCurAction := range cmdactions.ACTIONS {
		fmt.Printf("\t%s : %s\n", vCurActionName, vCurAction.GetDescription())
	}
	os.Exit(1)
}

func main() {

	if len(os.Args) == 1 {
		usage()
	}

	vSelectedActionName := os.Args[1]

	diagnostic.LogInfo("main", "Selected action %s", vSelectedActionName)
	vSelectedAction, vActionFound := cmdactions.ACTIONS[vSelectedActionName]

	if vActionFound == false {
		diagnostic.LogFatal("main", "action %s not found", nil, vSelectedActionName)
	}

	diagnostic.LogInfo("main", "Executing initialization...")
	vPreInitializer, vIsPreInitalizable := vSelectedAction.(cmdactions.PreInitializable)

	if vIsPreInitalizable {
		vPreInitError := vPreInitializer.PreInit()
		diagnostic.LogFatalIfError(vPreInitError, "main", "the following error is occurred during pre inizialition")

	}

	vActionArgs := []string{}
	if len(os.Args) > 2 {
		vActionArgs = os.Args[2:]
	}
	vParseError := vSelectedAction.ParseCmdLine(vActionArgs)
	diagnostic.LogFatalIfError(vParseError, "main", " An error occurred while parsing parameters")

	vInitializer, vIsInitalizable := vSelectedAction.(cmdactions.Initializable)

	if vIsInitalizable {
		vInitError := vInitializer.Init()
		diagnostic.LogFatalIfError(vInitError, "main", "An error occurre during the initialization")
	}

	diagnostic.LogInfo("main", "%s started", vSelectedActionName)
	vExecutionError := vSelectedAction.Execute()
	diagnostic.LogFatalIfError(vExecutionError, "main", "An error occurred while executing the action")

	vPostExecution, vIsPostExecute := vSelectedAction.(cmdactions.PostExecution)

	if vIsPostExecute {
		vPostExecuteError := vPostExecution.PostExecution()

		if vPostExecuteError != nil {
			diagnostic.LogFatalIfError(vExecutionError, "main", "the following error is occurred during action post execute")
		}
	}

	diagnostic.LogInfo("main", "%s succeded", vSelectedActionName)


	vCloser, vIsCloser := vSelectedAction.(io.Closer)
	if vIsCloser {
		vCloseError:=vCloser.Close()
		if vCloseError != nil {
			diagnostic.LogWarning("main", "an error occurred while closing action", vCloseError)
		}	
	}
}
