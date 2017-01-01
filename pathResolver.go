package downthedrive

import (
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
	"net/url"
	"strconv"
	"strings"
)

//PathResolverFunc function responsible to translate item to local path
//Parameters:
//	*onedriveclient.OneDriveItem = one drive item
//	string = local base path
//	string = base drive path
//Returns:
//	string = path of the drive item locally, empty in the item is not stored locally
//	error = nil if succeded, otherwise the error occurred
type PathResolverFunc func(*onedriveclient.OneDriveItem, string, string) (string, error)

//PathResolverByItemIdAndDate store files locally by item id and date (for a very basic versioning)
func PathResolverByItemIdAndDate(pOneDriveItem *onedriveclient.OneDriveItem, pBasePathLocal string, pBasePathDrive string) (string, error) {

	if pOneDriveItem.IsFolder() {
		return "", nil
	}

	return pBasePathLocal + "/" + pOneDriveItem.Id + "__" + strconv.FormatInt(pOneDriveItem.LastModifiedDateTime.Unix(), 10), nil
}

//BasicPathResolver file maintains the same path locally
func BasicPathResolver(pOneDriveItem *onedriveclient.OneDriveItem, pBasePathLocal string, pBasePathDrive string) (string, error) {

	vString, vError := url.QueryUnescape(pOneDriveItem.GetFullOneDrivePath())
	if vError != nil {
		return "", diagnostic.NewError("Error parsing  path of onedriveitem %s", vError, pOneDriveItem)
	}
	return strings.Replace(vString, pBasePathDrive, pBasePathLocal, -1), nil
}
