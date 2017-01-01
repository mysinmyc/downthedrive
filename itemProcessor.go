package downthedrive

import (
	"github.com/mysinmyc/gocommons/concurrent"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
)

//OneDriveItemDispatcherFunc signature of dispatcher method
//Parameters:
//  pContext = DownTheDrive context 
//  pItem = item to process
//  pWorkerId = id of the current worker
//Returns:
//  nil in case of success otherwise the error
type OneDriveItemDispatcherFunc func(pContext *DownTheDriveContext, pItem *onedriveclient.OneDriveItem, pWorkerId int) error

//ItemProcessor is an object for item recurring operations
type ItemProcessor struct {

	//Items dispatcher
	dispatcher *concurrent.Dispatcher

	//parent
	Parent *DownTheDrive

	contexts []*DownTheDriveContext

	oneDriveItemDispatcherFunc OneDriveItemDispatcherFunc

	indexStrategy IndexStrategy
}

//NewItemProcessor create a new instance of item processor
func (vSelf *DownTheDrive) NewItemProcessor(pOneDriveItemDispatcherFunc OneDriveItemDispatcherFunc, pIndexStategy IndexStrategy, pWorkers int) (vRisItemProcessor *ItemProcessor, vError error) {

	vContexts := make([]*DownTheDriveContext, pWorkers)
	vRis := &ItemProcessor{Parent: vSelf, contexts: vContexts, oneDriveItemDispatcherFunc: pOneDriveItemDispatcherFunc, indexStrategy: pIndexStategy}

	for vCnt := 0; vCnt < pWorkers; vCnt++ {
		vCurContext, vError := vSelf.NewContext()
		if vError != nil {
			return nil, diagnostic.NewError("error creating context", vError)
		}
		vRis.contexts[vCnt] = vCurContext
	}

	vRis.dispatcher = concurrent.NewDispatcher(vRis.consumerFunction, 10)

	return vRis, nil
}

//Enqueue an item
//Parameters:
// pItem = item to process
func (vSelf *ItemProcessor) Enqueue(pItem interface{}) {

	vSelf.dispatcher.Enqueue(pItem)
}

func (vSelf *ItemProcessor) Process(pItem interface{}) error {
	vError := vSelf.dispatcher.Start(len(vSelf.contexts))
	if vError != nil {
		return diagnostic.NewError("Error starting dispatcher", vError)
	}
	vSelf.dispatcher.Enqueue(pItem)
	vSelf.dispatcher.WaitForCompletition()

	if vSelf.dispatcher.IsSucceded() == false {
		return diagnostic.NewError("Processor encountered errors processing some items", nil)
	}
	return nil
}

func (vSelf *ItemProcessor) consumerFunction(pDispatcher *concurrent.Dispatcher, pWorkerId int, pItem interface{}, pWorkerLocals concurrent.WorkerLocals) error {

	if diagnostic.IsLogTrace() {

		diagnostic.LogTrace("ItemProcessor.consumerFunction", "processing %s", pItem)
	}
	vItem, vOk := pItem.(*onedriveclient.OneDriveItem)
	if vOk == false || (vItem.IsFolder() && vItem.Children == nil && vItem.Folder.ChildCount != 0) {
		vCurItemNew, vGetItemError := vSelf.Parent.GetItem(pItem, true, vSelf.indexStrategy, vSelf.contexts[pWorkerId])
		vItem = vCurItemNew
		if vGetItemError != nil {
			return diagnostic.NewError("Error getting item %s", vGetItemError, pItem)
		}
	}

	vDispatchingError := vSelf.oneDriveItemDispatcherFunc(vSelf.contexts[pWorkerId], vItem, pWorkerId)

	if vDispatchingError != nil {
		return diagnostic.NewError("Error executing dispatching of item %s", vDispatchingError, pItem)
	}

	if vItem.IsFolder() == false {
		return nil
	}

	for _, vCurChild := range vItem.Children {
		vSelf.Enqueue(vCurChild)
	}

	return nil
}

type ItemProcessorWorkerConsumerFunc func(pContext *DownTheDriveContext, pWorkerId int) error

func (vSelf *ItemProcessor) ForEachWorker(pConsumerFunc ItemProcessorWorkerConsumerFunc) error {
	for vCnt, vCurContext := range vSelf.contexts {
		vError := pConsumerFunc(vCurContext,vCnt)
		if vError != nil {
			return diagnostic.NewError("An error occurred while consuming contextg %d", vError,vCnt)
		}
	}
	return nil
}

func (vSelf *ItemProcessor) WaitForCompletition() {
	vSelf.dispatcher.WaitForCompletition()
}

func (vSelf *ItemProcessor) IsSucceded() bool {
	return vSelf.dispatcher.IsSucceded()
}


func (vSelf *ItemProcessor) Close() error {
	return vSelf.ForEachWorker(func(pContext *DownTheDriveContext, pWorkerId int) error {
		return	pContext.Close()
		})
}
