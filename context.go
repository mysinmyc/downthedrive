package downthedrive 

import (
	"github.com/mysinmyc/onedriveclient"
	"github.com/mysinmyc/gocommons/diagnostic"
)

//DownTheDriveContext: holder for all the local objects needed by downthedrive
type DownTheDriveContext struct {
	parent *DownTheDrive
	db *DownTheDriveDb
	oneDriveClient *onedriveclient.OneDriveClient
}

//NewContext: return a new context
func (vSelf *DownTheDrive) NewContext() (*DownTheDriveContext, error) {
	return &DownTheDriveContext{parent:vSelf},nil
}

//GetDefaultContext: return a context
func (vSelf *DownTheDrive) GetDefaultContext() (*DownTheDriveContext, error) {
	if vSelf.defaultContext != nil {
		return vSelf.defaultContext,nil	
	}
	vDefaultContext, vDefaultContextError:= vSelf.NewContext()
	if vDefaultContextError != nil {
		return nil, diagnostic.NewError("An error occurred while creating default context ",vDefaultContextError)
	}
	vSelf.defaultContext = vDefaultContext
	return vDefaultContext,nil
}

//GetOneDriveClient: return a onedrive object instance for the current context
func (vSelf *DownTheDriveContext) GetOneDriveClient() (*onedriveclient.OneDriveClient, error) {

	if vSelf.oneDriveClient != nil {
		return vSelf.oneDriveClient,nil
	}

        vOneDriveClient := onedriveclient.NewOneDriveClient()
        vOneDriveClient.SetAuthenticationProvider(vSelf.parent)
	vSelf.oneDriveClient = vOneDriveClient
        return vOneDriveClient, nil
	
}

//GetDb: return a db object instance for the current context
func (vSelf *DownTheDriveContext) GetDb() (*DownTheDriveDb, error) {
	
	if vSelf.db != nil {
		return vSelf.db,nil
	}
       
	vDb, vDbError := NewDb(vSelf.parent.config.Db.Driver, vSelf.parent.config.Db.DataSource)
        if vDbError != nil {
                return nil, diagnostic.NewError("Failed to initialize database ", vDbError)
        }
	vSelf.db=vDb
        return  vDb,nil
}


func (vSelf *DownTheDriveContext) Close() error {

	if vSelf.db!=nil {
		vDbCloseError:= vSelf.db.Close()	

		if vDbCloseError != nil {
			return diagnostic.NewError("Error closing db from context",vDbCloseError)
		}
	}
	return nil
}
