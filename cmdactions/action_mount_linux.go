package cmdactions

import (
	"fmt"
	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/downthedrive/fs"
	"github.com/mysinmyc/gocommons/diagnostic"
)

type ActionMount struct {
	genericAction
	mountPoint  string
	allowOther  bool
}

func (vSelf *ActionMount) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()

	vFlagSet.StringVar(&vSelf.mountPoint, PARAMETER_MOUNTPOINT, "", "MountPoint")
	vFlagSet.BoolVar(&vSelf.allowOther, PARAMETER_ALLOWOTHER, false, "Allow other to access the filesystem")
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

func (vSelf *ActionMount) Execute() error {

	vDownTheDrive, vDownTheDriveError:= vSelf.GetDownTheDrive()
	if vDownTheDriveError != nil {
		return diagnostic.NewError("Error getting downthedrive", vDownTheDriveError)
	}

	vContext,vContextError:=vDownTheDrive.GetDefaultContext()
	if vContextError != nil {
		return diagnostic.NewError("Error getting default context", vContextError)
	}

	vOneDriveBaseItemPath, vOneDriveBaseItemPathError:= vDownTheDrive.GetItem(vSelf.drivePath,false,downthedrive.IndexStrategy_Default, vContext)
	if vOneDriveBaseItemPathError !=nil {
		return diagnostic.NewError("error getting onedrive path %s", vOneDriveBaseItemPathError, vSelf.drivePath)
	}


	vLocalStore, vLocalStoreError:= vSelf.GetLocalStore()
	if vLocalStoreError != nil {
		return diagnostic.NewError("Error getting local store", vLocalStoreError)
	}

	vDownTheDriveFs, vDownTheDriveFsError := fs.NewDownTheDriveFs(vDownTheDrive, vOneDriveBaseItemPath, vLocalStore)
	if vDownTheDriveFsError != nil {
		return diagnostic.NewError("An error occurred while creating mounting object for fs", vDownTheDriveFsError)
	}
	defer vDownTheDriveFs.Close()

	vMountError := vDownTheDriveFs.Mount(vSelf.mountPoint, vSelf.allowOther)

	if vMountError != nil {
		return diagnostic.NewError("failed to mount filesystem", vMountError)
	}
	
	return nil
}

func (vSelf *ActionMount) GetDescription() string {
	return "Mount the drive locally"
}
