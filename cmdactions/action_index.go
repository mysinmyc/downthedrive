package cmdactions

import (
	"fmt"	
	"github.com/mysinmyc/gocommons/diagnostic"
)

type ActionIndex struct {
	genericAction
	drivePath string
	threads   int
}

func (vSelf *ActionIndex) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()

	vFlagSet.IntVar(&vSelf.threads, PARAMETER_INDEXTHREADS, 4, "Number of indexing threads")
	vParseError := vFlagSet.Parse(pArgs)

	if vParseError != nil {
		return vParseError
	}

	vMandatoryParametersError := vSelf.CheckMandatoryParameters()
	if vMandatoryParametersError != nil {
		return vMandatoryParametersError
	}

	if vSelf.drivePath == "" {
		return fmt.Errorf("missing %s parameter ", PARAMETER_DRIVEPATH)
	}

	return nil
}

func (vSelf *ActionIndex) Execute() error {

	vDownTheDrive, vDownTheDriveError:= vSelf.GetDownTheDrive()
        if vDownTheDriveError != nil {
                return diagnostic.NewError("Error getting downthedrive", vDownTheDriveError)
        }

	vIndexer, vIndexerError := vDownTheDrive.NewIndexer(vSelf.threads)
	if vIndexerError != nil {
		return diagnostic.NewError("An error occurred while creating indexer", vIndexerError)
	}

	vStartIndexingError := vIndexer.StartIndexing()
	if vStartIndexingError != nil {
		return diagnostic.NewError("An error occurred while creating indexer", vStartIndexingError)
	}

	vIndexDrivePathError := vIndexer.IndexDrivePath(vSelf.drivePath)
	if vIndexDrivePathError != nil {
		return vIndexDrivePathError
	}

	vIndexer.WaitForCompletition()

	diagnostic.LogInfo("ActionIndex.Execute", "Drive indexing completed: path %s item indexed %d", vSelf.drivePath,vIndexer.CountIndexedItems())

	if vIndexer.IsSucceded() ==false {
		return diagnostic.NewError("Some errors occurred during indexing ",nil)
	}
	return nil
}

func (vSelf *ActionIndex) GetDescription() string {
	return "Index the content of the drive inside the local database"
}
