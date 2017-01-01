package cmdactions

import (
	"github.com/mysinmyc/gocommons/diagnostic"
)

type ActionDownload struct {
	genericAction
	drivePath       string
	localPath       string
	downloadThreads int
}

func (vSelf *ActionDownload) ParseCmdLine(pArgs []string) error {

	vFlagSet := vSelf.BuildFlagSet()
	vFlagSet.IntVar(&vSelf.downloadThreads, PARAMETER_DOWNLOADTHREADS, 4, "Number of download threads")
	vParseError := vFlagSet.Parse(pArgs)

	if vParseError != nil {
		return vParseError
	}

	vMandatoryParametersError := vSelf.CheckMandatoryParameters()
	if vMandatoryParametersError != nil {
		return vMandatoryParametersError
	}

	return nil
}

func (vSelf *ActionDownload) Execute() error {

	vDownTheDrive, vDownTheDriveError:= vSelf.GetDownTheDrive()
        if vDownTheDriveError != nil {
                return diagnostic.NewError("Error getting downthedrive", vDownTheDriveError)
        }

        vLocalStore, vLocalStoreError:= vSelf.GetLocalStore()
        if vLocalStoreError != nil {
                return diagnostic.NewError("Error getting local store", vLocalStoreError)
        }

	vDownloader, vError := vDownTheDrive.NewDownloader(vLocalStore, vSelf.downloadThreads)
	if vError != nil {
		return diagnostic.NewError("An error occurred while creating downloader",vError)
	}

	vError = vDownloader.Download()
	if vError != nil {
		return diagnostic.NewError("An error occurred during download",vError)
	}
	
	vError = vDownloader.Close()
	if vError != nil {
		return diagnostic.NewError("An error occurred while closing downloader",vError)
	}
	return nil
}

func (vSelf *ActionDownload) GetDescription() string {
	return "Download from drive"
}
