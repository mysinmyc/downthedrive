package httpd

import (
	"net/http"
	"net/url"

	"github.com/mysinmyc/downthedrive"
	"github.com/mysinmyc/onedriveclient"
)

type ItemHandlerBase struct {
	context        *downthedrive.DownTheDriveContext
	downTheDrive   *downthedrive.DownTheDrive
}

type ItemForTemplate struct {
	onedriveclient.OneDriveItem
	IsFolder bool
}

func (vSelf *ItemHandlerBase) getRequestedItem(pRequest *http.Request) (vRisItem *onedriveclient.OneDriveItem, vRisError error) {

	vId, _ := url.QueryUnescape(pRequest.FormValue("id"))
	if vId != "" {
		return vSelf.downTheDrive.GetItem(&onedriveclient.OneDriveItem{Id: vId}, true, downthedrive.IndexStrategy_Default, vSelf.context)
	}
	return vSelf.downTheDrive.GetItem(pRequest.FormValue("drivePath"), true, downthedrive.IndexStrategy_Default, vSelf.context)
}
