package cmdactions

import (
	"fmt"

	"github.com/mysinmyc/gocommons/diagnostic"
)

type ActionHelp struct {
	actionName string
}

func (vSelf *ActionHelp) ParseCmdLine(pArgs []string) error {

	if len(pArgs) < 1 {
		return fmt.Errorf("No action specified")
	}
	vSelf.actionName = pArgs[0]
	return nil
}

func (vSelf *ActionHelp) Init() error {
	return nil
}

func (vSelf *ActionHelp) PostExecute() error {
	return nil
}

func (vSelf *ActionHelp) Execute() error {

	vAction, vFound := ACTIONS[vSelf.actionName]

	if vFound == false {
		return fmt.Errorf("Action %s doesn't exists", vSelf.actionName)
	}

	vPreInitializer, vIsPreInitalizable := vAction.(PreInitializable)

	if vIsPreInitalizable {
		vPreInitError := vPreInitializer.PreInit()

		if vPreInitError != nil {
			return diagnostic.NewError("Error during  target action preinitialization ", vPreInitError)
		}
	}

	fmt.Printf("\nACTION %s:\n\n", vSelf.actionName)
	vAction.PrintHelp()
	return nil
}

func (vSelf *ActionHelp) GetDescription() string {
	return "Print the help of a specific action"
}

func (vSelf *ActionHelp) PrintHelp() {

	fmt.Printf("PARAMETERS:\n\t{actionName}: name of the action \n")
}
