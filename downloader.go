package downthedrive

import (
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
)

//Downloader: allows to download items content from drive
type Downloader struct {
	downTheDrive  *DownTheDrive
	itemProcessor *ItemProcessor
	localStore    *LocalStore
}


//NewDownLoader: create a new instance of downloader
//Parameters:
// pDrivePath = path on the drive to download ( "/" = root of the drive)
// pLocalPath = local path where to downlaod item contents
// pWorkers = number of concurrent workers
//Returns:
// *Downloader = instance of the downloader
// error = nil if succeded otherwise the error
func (vSelf *DownTheDrive) NewDownloader(pLocalStore *LocalStore,pWorkers int) (vRisDownloader *Downloader, vRisError error) {

	vRis := &Downloader{downTheDrive: vSelf, localStore: pLocalStore}

	vItemProcessor, vError := vSelf.NewItemProcessor(vRis.downloadItemRecursively, IndexStrategy_Default, pWorkers)

	if vError != nil {
		return nil, diagnostic.NewError("Error instantiating item processor", vError)
	}

	vRis.itemProcessor = vItemProcessor
	return vRis, nil
}

func (vSelf *Downloader) downloadItemRecursively(pContext *DownTheDriveContext, pItem *onedriveclient.OneDriveItem, pWorker int) error {

	diagnostic.LogDebug("Downloader.downloadItemRecursively", "<%d> %s, Folder: %s, File:%s <%s>\n\n", pWorker, pItem.Name, pItem.IsFolder(), pItem.IsFile(), pItem.CreatedDateTime)

	_, vError := vSelf.localStore.Get(pItem, pContext)

	return vError
}

func (vSelf *Downloader) Download() error {
	diagnostic.LogInfo("Downloader.Download", "Download started...")
	vError := vSelf.itemProcessor.Process(vSelf.localStore.GetOneDriveBaseItemPath())
	if vError ==nil {
		diagnostic.LogInfo("Downloader.Download","Download succeded")
	} else {
		diagnostic.LogWarning("Downloader.Download","Download failed",vError)
		return diagnostic.NewError("An error occurred downloading items", vError)
	}


	return nil
}


func (vSelf *Downloader) Close() error {
	if vSelf.localStore != nil {
		return vSelf.localStore.Close() 
	}
	return nil
}
