package fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
)

type DownTheDriveFs struct {
	downTheDrive   *downthedrive.DownTheDrive
	context        *downthedrive.DownTheDriveContext
	localStore     *downthedrive.LocalStore
}

//NewDownTheDriveFs: create a new filesystem object 
//Parameters:
// pDownTheDrive = DownTheDrive instance
// pOneDriveBaseItemPath = base item path on onedrive
//
func NewDownTheDriveFs(pDownTheDrive *downthedrive.DownTheDrive, pOneDriveBaseItemPath *onedriveclient.OneDriveItem, pLocalStore *downthedrive.LocalStore) (*DownTheDriveFs, error) {

	vContext, vContextError := pDownTheDrive.NewContext()
	if vContextError != nil {
		return nil, diagnostic.NewError("Error creating context", vContextError)
	}

	return &DownTheDriveFs{downTheDrive: pDownTheDrive, context: vContext,  localStore: pLocalStore}, nil
}

//Unmount: unmount a filesystem
//Parameters:
// pMountPoint = mountpoint
//Returns: nil if succeceded otherwise the error
func Unmount(pMountPoint string) error {
	vUnmountError := fuse.Unmount(pMountPoint)
	if vUnmountError != nil {
		return diagnostic.NewError("An error occurred while unmounting onedrive filesystem mounted at %s", vUnmountError, pMountPoint)
	}
	return nil
}

//Mount: mount the filesystem
//Parameters:
// pMountPoint = mount point
// pAllowOther = if true allow others (and root) to access the filesystem
//Returns: nil if succeded othewise the error
func (vSelf *DownTheDriveFs) Mount(pMountPoint string,pAllowOther bool) error {

	vMountOptions := []fuse.MountOption{
		fuse.FSName("downthedrive"),
		fuse.Subtype("fs"),
		fuse.ReadOnly(),
		fuse.LocalVolume(),
		fuse.VolumeName("xx")}

	if pAllowOther {
		vMountOptions = append(vMountOptions, fuse.AllowOther())
	}

	vFileSystemConnection, vMountError := fuse.Mount(
		pMountPoint,
		vMountOptions...)
	if vMountError != nil {
		return diagnostic.NewError("An error occurred while mounting onedrive filesystem", vMountError)
	}
	defer vFileSystemConnection.Close()
	defer Unmount(pMountPoint)

	vServeError := fs.Serve(vFileSystemConnection, vSelf)
	if vServeError != nil {
		return diagnostic.NewError("An error occurred while serving onedrive filesystem", vServeError)
	}

	<-vFileSystemConnection.Ready
	if vFileSystemConnection.MountError != nil {
		return diagnostic.NewError("Mount failed", vFileSystemConnection.MountError)

	}
	return nil
}

func (vSelf *DownTheDriveFs) Root() (fs.Node, error) {
	return vSelf.getFsItem(vSelf.localStore.GetOneDriveBaseItemPath())
}

func (vSelf *DownTheDriveFs) getOneDriveClient() (*onedriveclient.OneDriveClient, error) {
	return vSelf.context.GetOneDriveClient()
}

func (vSelf *DownTheDriveFs) getFsItem(pOneDriveItem interface{}) (*OneDriveFsItem, error) {

	vOneDriveItem, vOneDriveItemError := vSelf.downTheDrive.GetItem(pOneDriveItem, true, downthedrive.IndexStrategy_Default, vSelf.context)

	if vOneDriveItemError != nil {
		return nil, diagnostic.NewError("Error getting onedrive item %s", vOneDriveItemError, pOneDriveItem)
	}

	return &OneDriveFsItem{parent: vSelf, oneDriveItem: vOneDriveItem}, nil

}


func (vSelf *DownTheDriveFs) Close() error {
	return vSelf.context.Close()	
}
