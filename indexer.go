package downthedrive

import (
	"github.com/mysinmyc/gocommons/concurrent"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
)

//Indexed: allows to index onedrive path into a database
type Indexer struct {
	started		    bool
	downTheDrive        *DownTheDrive
	itemsFromDriveProcessor       *ItemProcessor
	itemsToDbDispatcher *concurrent.Dispatcher
	indexItemChannel    chan *onedriveclient.OneDriveItem
	countIndexedItem    int
}

//NewIndexer: create a new instance of indexer
//Parameters:
//	pDriveWorkers = number of concurrent workers to the drive
//Returns:
// 	*Indexer = indexer object
//	error = nil if succeded otherwise the error occurred
func (vSelf *DownTheDrive) NewIndexer(pDriveWorkers int) (*Indexer, error) {

	vRis := &Indexer{downTheDrive: vSelf, indexItemChannel: make(chan *onedriveclient.OneDriveItem, 500)}

	vItemProcessor, vError := vSelf.NewItemProcessor(vRis.saveItem, IndexStrategy_DriveOnly, pDriveWorkers)

	if vError != nil {
		return nil, diagnostic.NewError("Error creating item processor ", vError)
	}

	vRis.itemsFromDriveProcessor = vItemProcessor

	vRis.itemsToDbDispatcher = concurrent.NewDispatcher(vRis.dbConsumerFunction, 100)
	vRis.itemsToDbDispatcher.WorkerLifeCycleHandlerFunc = vRis.indexDispatcherWorkerLifeCycleHandler
	return vRis, nil
}

func (vSelf *Indexer) indexDispatcherWorkerLifeCycleHandler(pDispatcher *concurrent.Dispatcher, pWorkerCnt int, pEvent concurrent.WorkerLifeCycleEvent, pWorkerLocals concurrent.WorkerLocals) (concurrent.WorkerLocals,error) {

	switch pEvent {
		case concurrent.WorkerLifeCycleEvent_Started:

			vContext,vContextError:= vSelf.downTheDrive.NewContext()
		
			if vContextError != nil {
				return nil, diagnostic.NewError("Error while initializing context", vContextError)
			}

			vDb,vDbError:= vContext.GetDb()
			if vDbError != nil {
				return nil, diagnostic.NewError("Error while initializing db from context", vContextError)
			}

			_,vBeginItemBulkError := vDb.BeginItemBulk()
			if vBeginItemBulkError != nil {
				return nil, diagnostic.NewError("Error during item bulk initialization", vBeginItemBulkError)
			}

			return vContext,nil
		case concurrent.WorkerLifeCycleEvent_Stopped:

			vContext, vIsContext := pWorkerLocals.(*DownTheDriveContext)
			if vIsContext ==false {
				return nil, diagnostic.NewError("i don't know why but workerlocals %#v is not a downthedrivecontext", nil,pWorkerLocals)
			}
		
			vDb,vDbError:= vContext.GetDb()
			if vDbError != nil {
				return nil, diagnostic.NewError("Error while getting db from context", vDbError)
			}

			vEndItemBulkError := vDb.EndItemBulk()
			if vEndItemBulkError != nil {
				return nil, diagnostic.NewError("Error ending item bulk", vEndItemBulkError)
			}

			vContextCloseError:= vContext.Close()
			if vContextCloseError != nil {	
				return nil, diagnostic.NewError("An error occurred while closing context",vContextCloseError)
			}
			
			return nil,nil
			
	}
	return nil,nil	
}

func (vSelf *Indexer) saveItem(pContext *DownTheDriveContext, pItem *onedriveclient.OneDriveItem, pWorker int) error {

	vSelf.itemsToDbDispatcher.Enqueue(pItem)
	return nil
}

func (vSelf *Indexer) dbConsumerFunction(pDispatcher *concurrent.Dispatcher, pWorkerId int, pItem interface{}, pWorkerLocals concurrent.WorkerLocals) error {


	vDb,_:=pWorkerLocals.(*DownTheDriveContext).GetDb()
	vError:=vDb.SaveItem(pItem.(*onedriveclient.OneDriveItem))

	if vError != nil {
		return diagnostic.NewError("Error saving item %s ", vError, pItem)
	}

	vSelf.countIndexedItem += 1
	if vSelf.countIndexedItem%500 == 0 {
		diagnostic.LogInfo("Indexer.dbConsumerFunction", "Item indexed %d ", vSelf.countIndexedItem)
	}
	return vError
}

//Build: index a drive path
//Parameters:
//	pPath = drive path to index
//Returns:
//	nil if succeded otherwise the error
func (vSelf *Indexer) IndexDrivePath(pPath string) error {

	var vError error
	vError = vSelf.IndexUpToRoot(pPath, nil)
	if vError != nil {
		return diagnostic.NewError("Failed to index to root ", vError)
	}

	vError = vSelf.itemsFromDriveProcessor.Process(pPath)
	if vError != nil {
		return diagnostic.NewError("Error processing path %s", vError, pPath)
	}
	
	return nil
}

//Index: index a single item into the database
//Parameters:
//	pItem = item to index
//	pContext = instance of DownTheDriveContext
//Returns:
//	error = nil if succeded otherwise the error
func (vSelf *Indexer) Index(pItem interface{}, pContext *DownTheDriveContext) error {
	_, vError := vSelf.index(pItem, pContext)
	return vError
}

func (vSelf *Indexer) index(pItem interface{}, pContext *DownTheDriveContext) (*onedriveclient.OneDriveItem, error) {
	
	vItem, vItemError := vSelf.downTheDrive.GetItem(pItem, false, IndexStrategy_DriveOnly, pContext)
	if vItemError != nil {
		return nil, diagnostic.NewError("failed to get item %s", vItemError, pItem)
	}
	vSelf.itemsToDbDispatcher.Enqueue(vItem)

	return vItem, nil
}

//IndexUpToRoot: index an item and parents up to root
//Parameters:
//	pItem = item to index
//	pContext = instance of DownTheDriveContext
//Returns:
//	nil if succeded otherwise the error
func (vSelf *Indexer) IndexUpToRoot(pItem interface{}, pContext *DownTheDriveContext) error {

	vItem, vItemError := vSelf.index(pItem, pContext)
	if vItemError != nil {
		return vItemError
	}
	if vItem.ParentReference != nil {
		return vSelf.IndexUpToRoot(vItem.ParentReference.Id, pContext)
	}

	return nil
}

func (vSelf *Indexer) StartIndexing() error {
	return vSelf.itemsToDbDispatcher.Start(1)
}

func (vSelf *Indexer) WaitForCompletition() {
	vSelf.itemsFromDriveProcessor.WaitForCompletition()
	vSelf.itemsToDbDispatcher.WaitForCompletition()
	vSelf.started=false
}

func (vSelf *Indexer) IsSucceded() bool {
	return vSelf.itemsFromDriveProcessor.IsSucceded() && vSelf.itemsToDbDispatcher.IsSucceded()
}

func (vSelf *Indexer) CountIndexedItems() int {
	return vSelf.countIndexedItem
}
