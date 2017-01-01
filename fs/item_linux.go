package fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
	"golang.org/x/net/context"
	"os"
	"sync"
)

type OneDriveFsItem struct {
	parent       *DownTheDriveFs
	oneDriveItem *onedriveclient.OneDriveItem
	mutex        sync.Mutex
}

func (vSelf *OneDriveFsItem) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mtime = vSelf.oneDriveItem.LastModifiedDateTime
	a.Uid = uint32(os.Getuid())
	a.Gid = uint32(os.Getgid())
	if vSelf.oneDriveItem.IsFolder() {
		a.Mode = os.ModeDir | 0555
	} else {
		a.Size = uint64(vSelf.oneDriveItem.SizeBytes)
		//a.Blocks = a.Size * 2
		a.Mode = 0444
	}
	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("OneDriveFsItem.Attr", "returned %#v attributes for item isFolder:%v %s", a, vSelf.oneDriveItem.IsFolder(), vSelf.oneDriveItem)
	}
	return nil
}

func (vSelf *OneDriveFsItem) getChild(pName string) (*OneDriveFsItem, error) {

	if vSelf.oneDriveItem.Children != nil {
		for _, vCurItem := range vSelf.oneDriveItem.Children {
			if vCurItem.Name == pName {
				return vSelf.parent.getFsItem(vCurItem)
			}
		}
	}

	return nil, nil
}

func (vSelf *OneDriveFsItem) Lookup(pContext context.Context, pName string) (fs.Node, error) {

	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("OneDriveFsItem.Lookup", "Lookup of %s child of %s", pName, vSelf.oneDriveItem)
	}
	vItem, vItemError := vSelf.getChild(pName)

	if vItemError != nil {
		return nil, diagnostic.NewError("error during lookup of %s in %s", vItemError, pName, vSelf)
	}
	if vItem != nil {
		return vItem, nil
	}
	return nil, fuse.ENOENT
}

func (vSelf *OneDriveFsItem) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("OneDriveFsItem.ReadDirAll", "Reading %s...", vSelf.oneDriveItem)
	}
	vRis := make([]fuse.Dirent, 0, 10)

	if vSelf.oneDriveItem.Children != nil {
		for _, vCurItem := range vSelf.oneDriveItem.Children {
			vRis = append(vRis, fuse.Dirent{Name: vCurItem.Name})
		}
	}
	return vRis, nil
}

func (vSelf *OneDriveFsItem) Open(pContext context.Context, pRequest *fuse.OpenRequest, pResponse *fuse.OpenResponse) (fs.Handle, error) {

	if vSelf.oneDriveItem.IsFolder() {
		return vSelf, nil
	}

	vSelf.mutex.Lock()
	defer vSelf.mutex.Unlock()

	vFileReference, vFileReferenceError := vSelf.parent.localStore.Get(vSelf.oneDriveItem, vSelf.parent.context)
	if vFileReferenceError != nil {
		return nil, diagnostic.NewError("Failed to obtain file reference on the localstore for %s", vFileReferenceError, vSelf.oneDriveItem)
	}

	vFile, vFileError:=  vSelf.parent.localStore.Open(vFileReference)
	if vFileError != nil {
		return nil, diagnostic.NewError("Failed to open file reference on the localstore for %s", vFileError, vSelf.oneDriveItem)
	}

	return &FileHandle{file:vFile}, nil
}
