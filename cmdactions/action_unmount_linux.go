package cmdactions

import (
	"fmt"
	"github.com/mysinmyc/downthedrive/fs"
	"github.com/mysinmyc/gocommons/diagnostic"
)

type ActionUnmount struct {
	genericAction
	mountPoint string
}

func (vSelf *ActionUnmount) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()

	vFlagSet.StringVar(&vSelf.mountPoint, PARAMETER_MOUNTPOINT, "", "MountPoint")
	vParseError := vFlagSet.Parse(pArgs)

	if vParseError != nil {
		return vParseError
	}

	vMandatoryParametersError := vSelf.CheckMandatoryParameters()
	if vMandatoryParametersError != nil {
		return vMandatoryParametersError
	}

	if vSelf.mountPoint == "" {
		return fmt.Errorf("missing %s parameter ", PARAMETER_MOUNTPOINT)
	}

	return nil
}

func (vSelf *ActionUnmount) Execute() error {
	vUnmountError := fs.Unmount(vSelf.mountPoint)

	if vUnmountError != nil {
		return diagnostic.NewError("failed to unmount fs on %s", vUnmountError, vSelf.mountPoint)
	}
	return nil
}

func (vSelf *ActionUnmount) GetDescription() string {
	return "Ummount a local mounted drive"
}
