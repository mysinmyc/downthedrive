package httpd

import (
	"net/http"
	"net/url"
	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/gocommons/diagnostic"
)

type DownloadHandler struct {
	localStore     *downthedrive.LocalStore
	downTheDrive   *downthedrive.DownTheDrive
}

func RegisterDownloadHandler(pDownTheDrive *downthedrive.DownTheDrive, pLocalStore *downthedrive.LocalStore) error {

	http.Handle("/download", &DownloadHandler{downTheDrive: pDownTheDrive, localStore:pLocalStore})
	return nil

}

func (vSelf *DownloadHandler) ServeHTTP(pResponse http.ResponseWriter, pRequest *http.Request) {

	vContext,vContextError := vSelf.downTheDrive.GetDefaultContext()
	if vContextError !=nil {
		diagnostic.LogError("DownloadHandler.ServeHTTP", "Error getting context", vContextError)
		http.Error(pResponse, "Error getting context", 500)
	}

	vId, _ := url.QueryUnescape(pRequest.FormValue("id"))
	if vId == "" {
		diagnostic.LogError("DownloadHandler.ServeHTTP", "Received a request without id %s", nil, pRequest.URL)
		http.Error(pResponse, "Missing id", 500)
		return
	}

	vOneDriveItem, vOneDriveItemError:=vSelf.downTheDrive.GetItem(vId,false,downthedrive.IndexStrategy_Default,vContext)
	if vOneDriveItemError != nil {
		diagnostic.LogError("DownloadHandler.ServeHTTP", "Error getting onedriveitem", vOneDriveItemError)
		http.Error(pResponse, "Error getting onedriveitem", 500)
	}

	vContentType, _ := url.QueryUnescape(pRequest.FormValue("contentType"))
	if vContentType != "" {
		pResponse.Header().Add("content-type", vContentType)
	}

	vFileName, _ := url.QueryUnescape(pRequest.FormValue("fileName"))
	if vFileName == "" {
		vFileName = vId
	}

	pResponse.Header().Add("Content-Disposition", "inline; filename="+vFileName)

	vItemReference,vItemReferenceError:= vSelf.localStore.Get(vOneDriveItem,vContext)
	if vItemReferenceError != nil {
		diagnostic.LogError("DownloadHandler.ServeHTTP", "Error getting item reference from localstore", vItemReferenceError)
		http.Error(pResponse, "Error getting file reference from localStore", 500)
	}

	vWriteFileError := vSelf.localStore.WriteInto(vItemReference, pResponse)
	if vWriteFileError != nil {
		diagnostic.LogError("DownloadHandler.ServeHTTP", "Error writing file from localstore", vWriteFileError)
		http.Error(pResponse, "Error writing file from localstore", 500)
	}
}
