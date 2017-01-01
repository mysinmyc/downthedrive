package downthedrive

import (
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
	"time"
)

type IndexStrategy int

const (
	IndexStrategy_IndexOrDrive IndexStrategy = 0
	IndexStrategy_IndexOnly    IndexStrategy = iota
	IndexStrategy_DriveOnly    IndexStrategy = iota
	IndexStrategy_Default      IndexStrategy = iota
	
	ItemsMaxAge_None           time.Duration = -1	
	ItemsMaxAge_Unspecified    time.Duration = 0	
	ItemsMaxAge_Default        time.Duration = time.Hour * 24*15	
)

func (vSelf *DownTheDrive) SetDefaultIndexStrategy(pDefaultIndexStrategy IndexStrategy) {
	vSelf.config.IndexStrategy = pDefaultIndexStrategy
}

func (vSelf *DownTheDrive) HasItemsExpiry() bool {
	return vSelf.config.ItemsMaxAge > ItemsMaxAge_None
}

func (vSelf *DownTheDrive) GetItemsMaxAge() time.Duration{

	if vSelf.config.ItemsMaxAge == ItemsMaxAge_Unspecified {
		return ItemsMaxAge_Default
	}
	return vSelf.config.ItemsMaxAge
}

func (vSelf *DownTheDrive) GetItem(pItem interface{}, pExpandChilden bool, pIndexStrategy IndexStrategy, pContext *DownTheDriveContext) (vRisItem *onedriveclient.OneDriveItem, vRisError error) {

	var vContext *DownTheDriveContext
	if pContext==nil {
		var vContextError error
		vContext,vContextError = vSelf.GetDefaultContext()
		if vContextError != nil {
			return nil,diagnostic.NewError("Error while getting default context", vContextError)
		}
	} else {
		vContext = pContext
	}
	
	vIndexStrategy := pIndexStrategy
	if vIndexStrategy == IndexStrategy_Default {
		vIndexStrategy = vSelf.config.IndexStrategy
	}

	if vIndexStrategy != IndexStrategy_DriveOnly {

		diagnostic.LogDebug("DownTheDrive.GetItem", "Searching item in db %s", pItem)
		
		vDb,vDbError := vContext.GetDb()
		if vDbError != nil {
			diagnostic.NewError("Error getting db from context",vDbError)
		}
		
		vItemInDb, vItemInDbError := vDb.GetItem(pItem, pExpandChilden)

		if vItemInDbError != nil {
			return nil, diagnostic.NewError("Failed to load item form db %s ", vItemInDbError, pItem)
		}

		if vItemInDb != nil && pExpandChilden && pIndexStrategy != IndexStrategy_IndexOnly && vItemInDb.ChildrenLoaded() == false {
			diagnostic.LogDebug("DownTheDrive.GetItem", "Item in db does not contains all the children, reloading from drive")
			vItemInDb = nil
		}

		if vItemInDb != nil {
			if vSelf.HasItemsExpiry() {
				if time.Now().Sub(vItemInDb.LocalInfo.SnapShotDate) > vSelf.GetItemsMaxAge() {
					diagnostic.LogDebug("DownTheDrive.GetItem", "Item in db expired") 
					if vIndexStrategy != IndexStrategy_IndexOnly {
						diagnostic.LogDebug("DownTheDrive.GetItem", "Forcing item reload")
						vItemInDb = nil
					}
				}
			}
		}

		if vItemInDb != nil {
			diagnostic.LogDebug("DownTheDrive.GetItem", "Item loaded from db")
			return vItemInDb, nil
		}

		if vIndexStrategy == IndexStrategy_IndexOnly {
			return nil, diagnostic.NewError("Item not found in the index db: %s", nil, vItemInDb)
		}

	}

	diagnostic.LogDebug("DownTheDrive.GetItem", "Getting item from drive  %s...", pItem)

	vOneDriveClient,vOneDriveClientError := vContext.GetOneDriveClient()
	if vOneDriveClientError != nil {
		return nil, diagnostic.NewError("An error occurred getting onedriveclient from context",vOneDriveClientError)
	}
	vItem, vItemError := vOneDriveClient.GetItem(pItem, pExpandChilden)

	if vItemError != nil {
		return nil, diagnostic.NewError("Error getting item from drive  %s", vItemError, pItem)
	}

	if pIndexStrategy != IndexStrategy_DriveOnly {
		diagnostic.LogDebug("DownTheDrive.GetItem", "Saving item %s in db...", vItem.Id)
		vDb,vDbError := vContext.GetDb()
		if vDbError != nil {
			diagnostic.NewError("Error getting db from context",vDbError)
		}
		vSaveError := vDb.SaveItem(vItem)

		if vSaveError != nil {
			return nil, diagnostic.NewError("Error saving item in database item from drive  %s", vSaveError, pItem)
		}

		if pExpandChilden && vItem.Children != nil {
			var vStartedBulk bool

			if len(vItem.Children) > 10 {
				vStartedBulk, vDbError = vDb.BeginItemBulk()

				if vDbError != nil {
					return nil, diagnostic.NewError("Error starting bulk to save children", vDbError)
				}
			}

			for _, vCurChildren := range vItem.Children {
				vDbError := vDb.SaveItem(vCurChildren)
				if vDbError != nil {
					return nil, diagnostic.NewError("error saving children %s ", vDbError, vCurChildren)
				}
			}

			if vStartedBulk {
				vDbError := vDb.EndItemBulk()

				if vDbError != nil {
					return nil, diagnostic.NewError("failed to end bulk", vDbError)
				}
			}
		}
	}
	return vItem, nil
}

