package downthedrive

import (
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/gocommons/myfileutils"
	"github.com/mysinmyc/onedriveclient"
	"os"
	"io"
	"time"
)

//LocalStore: manager of locally stored files
type LocalStore struct {
	basePathLocal        string
	basePathDrive        string
	oneDriveBaseItemPath *onedriveclient.OneDriveItem
	pathResolverFunc     PathResolverFunc
	LocalFileCheckerFunc LocalFileCheckerFunc
}

//LocalStoreItemReference: reference to an item stored locally
type LocalStoreItemReference struct {
	isFolder      bool
	localFilePath string
}

//NewLocalStore: create a new local store
func NewLocalStore(pBasePathLocal string, pOneDriveBaseItemPath *onedriveclient.OneDriveItem, pPathResolverFunc PathResolverFunc) (*LocalStore, error) {

	if pOneDriveBaseItemPath.IsFolder() == false {
		return nil, diagnostic.NewError("One drive item %s specified as base path is not a folder", nil, pOneDriveBaseItemPath)
	}

	vRis := &LocalStore{basePathLocal: pBasePathLocal, basePathDrive: pOneDriveBaseItemPath.GetFullOneDrivePath(), oneDriveBaseItemPath: pOneDriveBaseItemPath}

	if pPathResolverFunc == nil {
		vRis.pathResolverFunc = BasicPathResolver
	} else {
		vRis.pathResolverFunc = pPathResolverFunc
	}
	vRis.LocalFileCheckerFunc = (&LocalFileChecker{CheckSize: true, CheckModDate: true}).CheckerFunc
	return vRis, nil
}

func (vSelf *LocalStore) GetOneDriveBaseItemPath() *onedriveclient.OneDriveItem {
	return vSelf.oneDriveBaseItemPath
}

func (vSelf *LocalStore) resolvePath(pOneDriveItem *onedriveclient.OneDriveItem) (string, error) {
	vRis, vError := vSelf.pathResolverFunc(pOneDriveItem, vSelf.basePathLocal, vSelf.basePathDrive)

	if vError != nil {
		return "", diagnostic.NewError("Error resolving path", vError)
	}
	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("LocalStore.resolvePath", "OneDriveItem %s resolved to local path %s by %#v", pOneDriveItem, vRis, vSelf)

	}
	return vRis, nil
}

func (vSelf *LocalStore) checkFile(pOneDriveItem *onedriveclient.OneDriveItem, pLocalFilePath string) (bool, error) {
	return vSelf.LocalFileCheckerFunc(pOneDriveItem, pLocalFilePath)
}

func (vSelf *LocalStore) Exists(pOneDriveItem *onedriveclient.OneDriveItem) (bool, error) {
	vLocalPath, vLocalPathError := vSelf.resolvePath(pOneDriveItem)
	if vLocalPathError != nil {
		return false, diagnostic.NewError("Error resolving path of %s", vLocalPathError, pOneDriveItem)
	}

	vIsPresentLocally, vIsPresentLocallyError := vSelf.checkFile(pOneDriveItem, vLocalPath)
	if vLocalPathError != nil {
		return false, diagnostic.NewError("Error checking %s %s", vIsPresentLocallyError, pOneDriveItem, vLocalPath)
	}

	return vIsPresentLocally, nil
}

//Get: store locally the content of the specified onedriveitem if not present then return a refrence
//Parameters:
//	pOneDriveItem = item to store
//	pDownTheDriveContext = DownTheDrive context
//Return:
//	LocalStoreItemReference = a reference to open the file
//	error = nil if succeded otherwise the error
func (vSelf *LocalStore) Get(pOneDriveItem *onedriveclient.OneDriveItem, pContext *DownTheDriveContext) (*LocalStoreItemReference, error) {

	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("LocalStore.Get", "Checking %s", pOneDriveItem)
	}
	vLocalPath, vLocalPathError := vSelf.resolvePath(pOneDriveItem)
	if vLocalPathError != nil {
		return nil, diagnostic.NewError("Error resolving path of %s", vLocalPathError, pOneDriveItem)
	}

	if vLocalPath == "" {
		return nil, nil
	}

	vIsPresentLocally, vIsPresentLocallyError := vSelf.checkFile(pOneDriveItem, vLocalPath)
	if vIsPresentLocallyError != nil {
		return nil, diagnostic.NewError("Error checking %s %s", vIsPresentLocallyError, pOneDriveItem, vLocalPath)
	}

	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("LocalStore.Get", "Local presence of file %s for item %s is %s", vLocalPath, pOneDriveItem, vIsPresentLocally)
	}

	if vIsPresentLocally == false {

		if pOneDriveItem.IsFolder() {
			vMkDirError := myfileutils.MkDir(vLocalPath)
			if vMkDirError != nil {
				return nil, diagnostic.NewError("Error creating directory for file %s", vMkDirError, vLocalPath)
			}

		} else {

			vMkDirError := myfileutils.MkDir(myfileutils.GetParentPath(vLocalPath))
			if vMkDirError != nil {
				return nil, diagnostic.NewError("Error creating directory for file %s", vMkDirError, vLocalPath)
			}

			vStagingFile, vFileError := os.Create(vLocalPath)
			if vFileError != nil {
				return nil, diagnostic.NewError("Error creating local file %s", vFileError, vLocalPath)
			}

			vOneDriveClient, vOneDriveClientError:=pContext.GetOneDriveClient()
			if vOneDriveClientError != nil {
				return nil,diagnostic.NewError("Error getting onedriveclient from context", vOneDriveClientError)
			}
			vDownloadError := vOneDriveClient.DownloadContentInto(pOneDriveItem, vStagingFile)
			if vDownloadError != nil {
				return nil, diagnostic.NewError("Error downloading onedrive item %s ", vDownloadError, pOneDriveItem)
			}

			vCloseError := vStagingFile.Close()
			if vCloseError != nil {
				return nil, diagnostic.NewError("Failed to close file", vCloseError)
			}

			vChTimesError := os.Chtimes(vLocalPath, time.Now(), pOneDriveItem.LastModifiedDateTime)
			if vChTimesError != nil {
				return nil, diagnostic.NewError("Failed to change file time", vChTimesError)
			}
		}
	}

	return &LocalStoreItemReference{localFilePath: vLocalPath, isFolder: pOneDriveItem.IsFolder()}, nil
}

//Open: open the file 
//Parameters:
//	*LocalStoreItemReference = item reference
//Returns:
//	*os.File = local file
//	error = nil if succeded otherwise the error
func (vSelf *LocalStore) Open(pItemReference *LocalStoreItemReference) (*os.File, error) {

	if pItemReference.isFolder {
		return nil, diagnostic.NewError("Cannot open folder %s", nil, pItemReference.localFilePath)
	}

	vFile, vOpenError := os.OpenFile(pItemReference.localFilePath, os.O_RDONLY, 0660)
	if vOpenError != nil {
		return nil, diagnostic.NewError("Error opening file in the localstore %s", vOpenError, pItemReference.localFilePath)
	}

	return vFile, nil
}

//WriteInto write the content of the item into a writer
//Parameters:
//	*LocalStoreItemReference = item reference
//	io.Writer = Destination
//Returns: nil if succeded otherwise the error
func (vSelf *LocalStore) WriteInto(pItemReference *LocalStoreItemReference, pDestination io.Writer ) error {

	vFile,vFileError:= vSelf.Open(pItemReference)
	if vFileError !=nil {
		return diagnostic.NewError("Failed to open file in localstore",vFileError)
	}

	defer vFile.Close()
	
	_,vCopyError:=io.Copy(pDestination,vFile)
	if vCopyError !=nil {
		return diagnostic.NewError("Failed to copy file content",vCopyError)
	}

	return nil
}

func (vSelf *LocalStore) Close() error {
	return nil
}
