package downthedrive

import (
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
	"os"
	"time"
)

//LocalFileCheckerFunc: function to check if a onedriveitem and a local file have the same content
//Parameters:
//	OneDriveItem = One Drive Item
//	string = local file path
//Returns:
//	bool=true if local file exists and is the same of remote file, otherwise false
//	error=error if something goes wrong
type LocalFileCheckerFunc func(pOneDriveItem *onedriveclient.OneDriveItem, pLocalFilePath string) (bool, error)

type LocalFileChecker struct {
	CheckSize    bool
	CheckModDate bool
}

func (vSelf *LocalFileChecker) CheckerFunc(pOneDriveItem *onedriveclient.OneDriveItem, pLocalFilePath string) (bool, error) {

	vFileInfo, vFileInfoError := os.Stat(pLocalFilePath)

	if os.IsNotExist(vFileInfoError) {
		return false, nil
	}

	if vFileInfoError != nil {
		return false, diagnostic.NewError("An error occurred while stats file %s", vFileInfoError, vFileInfo)
	}

	if pOneDriveItem.IsFolder() == false {
		if vSelf.CheckSize {
			if pOneDriveItem.SizeBytes != vFileInfo.Size() {
				return false, nil
			}
		}

		if vSelf.CheckModDate {

			if pOneDriveItem.LastModifiedDateTime.Truncate(time.Millisecond).Equal(vFileInfo.ModTime().Truncate(time.Millisecond)) == false {
				return false, nil
			}
		}
	}
	return true, nil
}
