package downthedrive

import (
	"fmt"
	"net/url"
	"strings"

	"time"

	"database/sql"

	"github.com/mysinmyc/gocommons/db"
	"github.com/mysinmyc/gocommons/diagnostic"
	"github.com/mysinmyc/onedriveclient"
)

const (
	TABLE_ITEMS                 = "items"
	FIELD_ITEMS_ID              = "id"
	FIELD_ITEMS_PATH            = "path"
	FIELD_ITEMS_NAME            = "name"
	FIELD_ITEMS_PARENTID        = "parent_id"
	FIELD_ITEMS_PARENTPATH      = "parent_path"
	FIELD_ITEMS_SNAPSHOT_DATE     = "snapshot_date"
	FIELD_ITEMS_CREATETIMESTAMP = "create_timestamp"
	FIELD_ITEMS_UPDATETIMESTAMP = "update_timestamp"
	FIELD_ITEMS_FOLDERCHILDREN  = "folder_children"
	FIELD_ITEMS_SIZEBYTES       = "size_bytes"
	FIELD_ITEMS_HASH_SHA1       = "hash_sha1"

	DDL_ITEMS_SQLITE = "create table if not exists " + TABLE_ITEMS + " (" +
		FIELD_ITEMS_ID + " text PRIMARY KEY " +
		"," + FIELD_ITEMS_NAME + " text " +
		"," + FIELD_ITEMS_SNAPSHOT_DATE + " timestamp " +
		"," + FIELD_ITEMS_CREATETIMESTAMP + " timestamp " +
		"," + FIELD_ITEMS_UPDATETIMESTAMP + " timestamp " +
		"," + FIELD_ITEMS_FOLDERCHILDREN + " integer " +
		"," + FIELD_ITEMS_SIZEBYTES + " integer " +
		"," + FIELD_ITEMS_PATH + " text " +
		"," + FIELD_ITEMS_PARENTID + " text " +
		"," + FIELD_ITEMS_PARENTPATH + " text" +
		"," + FIELD_ITEMS_HASH_SHA1 + " text" +
		")"

	DDL_ITEMS_MYSQL = "create table if not exists " + TABLE_ITEMS + " (" +
		FIELD_ITEMS_ID + " varchar(700) PRIMARY KEY " +
		"," + FIELD_ITEMS_NAME + " varchar(1000) " +
		"," + FIELD_ITEMS_SNAPSHOT_DATE + " timestamp " +
		"," + FIELD_ITEMS_CREATETIMESTAMP + " timestamp " +
		"," + FIELD_ITEMS_UPDATETIMESTAMP + " timestamp " +
		"," + FIELD_ITEMS_FOLDERCHILDREN + " integer " +
		"," + FIELD_ITEMS_SIZEBYTES + " integer " +
		"," + FIELD_ITEMS_PATH + " varchar(10000) " +
		"," + FIELD_ITEMS_PARENTID + " varchar(700) " +
		"," + FIELD_ITEMS_PARENTPATH + " varchar(10000) " +
		"," + FIELD_ITEMS_HASH_SHA1 + " varchar(200) " +
		")"
)

var (
	FIELDS_ITEMS = []string{FIELD_ITEMS_ID, FIELD_ITEMS_NAME, FIELD_ITEMS_FOLDERCHILDREN, FIELD_ITEMS_SNAPSHOT_DATE, FIELD_ITEMS_CREATETIMESTAMP, FIELD_ITEMS_UPDATETIMESTAMP, FIELD_ITEMS_PATH, FIELD_ITEMS_PARENTID, FIELD_ITEMS_PARENTPATH, FIELD_ITEMS_SIZEBYTES, FIELD_ITEMS_HASH_SHA1}
)

func (vSelf *DownTheDriveDb) BeginItemBulk() (vRisBeginned bool, vRisError error) {

	vInitError:= vSelf.init() 
	if vInitError != nil {
		return false, diagnostic.NewError("initialization failed", vInitError)	
	}
	vBeginBulk,vBeginBulkError:= vSelf.itemInsert.BeginBulk(db.BulkOptions{})
	if vBeginBulkError != nil {
		return false, diagnostic.NewError("failed to begin bulk ", vBeginBulkError)
	}
	return vBeginBulk,vBeginBulkError
}

func (vSelf *DownTheDriveDb) CommitItemBulk() error {

	if vSelf.itemInsert == nil {
		return nil
	}

	return vSelf.itemInsert.Commit()
}

func (vSelf *DownTheDriveDb) EndItemBulk() error {
	if vSelf.itemInsert == nil {
		return nil
	}

	return vSelf.itemInsert.EndBulk()
}

func normalizePath(pDrivePath string) (string, error) {

	vString, vError := url.QueryUnescape(pDrivePath)
	if vError != nil {
		return "", diagnostic.NewError("error unescaping patch %s", vError, pDrivePath)
	}
	return strings.Replace(vString, "/drive/root:", "", -1) + "/", nil
}


func (vSelf *DownTheDriveDb) init() error {

	if vSelf.itemInsert == nil {
		vItemInsert, vItemInsertError := vSelf.CreateInsert(TABLE_ITEMS, FIELDS_ITEMS, db.InsertOptions{Replace: true})

		if vItemInsertError != nil {
			return diagnostic.NewError("Error while create item insert", vItemInsertError)
		}
		vSelf.itemInsert = vItemInsert
	}

	return nil
}

func (vSelf *DownTheDriveDb) SaveItem(pItem *onedriveclient.OneDriveItem) error {

	vInitError:= vSelf.init() 
	if vInitError != nil {
		return diagnostic.NewError("initialization failed", vInitError)	
	}
	vData := make([]interface{}, len(FIELDS_ITEMS))

	vData[0] = pItem.Id
	vData[1] = pItem.Name

	if pItem.Folder == nil {
		vData[2] = -1
	} else {
		vData[2] = pItem.Folder.ChildCount
	}

	vData[3] = time.Now()
	vData[4] = pItem.CreatedDateTime
	vData[5] = pItem.LastModifiedDateTime

	if pItem.ParentReference == nil {
		vData[6] = "/"
	} else {

		vNormalizedPath, vNormalizedPathError := normalizePath(pItem.ParentReference.Path)

		if vNormalizedPathError != nil {
			return diagnostic.NewError("failed to normalize path of item %s", vNormalizedPathError, pItem)
		}

		vData[6] = vNormalizedPath
		vData[7] = pItem.ParentReference.Id
		vData[8] = pItem.ParentReference.Path

	}
	vData[9] = pItem.SizeBytes

	if pItem.File != nil {
		vData[10] = pItem.File.Hashes.Sha1Hash
	}

	var vDbError error

	_, vDbError = vSelf.itemInsert.Exec(vData...)

	if vDbError != nil {
		return diagnostic.NewError("Error saving item %s in db ", vDbError, pItem)
	}
	return nil
}

func (vSelf *DownTheDriveDb) GetItemsByConditions(pConditions string) (vRisItems []*onedriveclient.OneDriveItem, vRisError error) {

	vStatement := "select " + strings.Join(FIELDS_ITEMS, ",") + " from " + TABLE_ITEMS + " where " + pConditions
	if diagnostic.IsLogTrace() {
		diagnostic.LogTrace("DownTheDriveDb.GetItemsByConditions", vStatement)
	}

	vRows, vDbError := vSelf.Query(vStatement)

	if vDbError != nil {
		return nil, diagnostic.NewError("Failed to search item %s in db ", vDbError, vDbError)
	}
	defer vRows.Close()
	vRis := make([]*onedriveclient.OneDriveItem, 0, 10)

	for vRows.Next() {

		vCurItem := &onedriveclient.OneDriveItem{}

		var vParentId sql.NullString
		var vParentPath sql.NullString
		var vFolderChildren sql.NullInt64
		var vSha1Hash sql.NullString
		var vPath string
		vScanError := vRows.Scan(&vCurItem.Id, &vCurItem.Name, &vFolderChildren, &vCurItem.LocalInfo.SnapShotDate, &vCurItem.CreatedDateTime, &vCurItem.LastModifiedDateTime, &vPath, &vParentId, &vParentPath, &vCurItem.SizeBytes, &vSha1Hash)
		//vData:=make([]interface{},len(vColumns))
		//vScanError := vRows.Scan(vData...)
		if vScanError != nil {
			return nil, diagnostic.NewError("Error parsing results ", vScanError)
		}

		if vParentId.Valid {
			vCurItem.ParentReference = &onedriveclient.ParentReference{Id: vParentId.String}

			if vParentPath.Valid {
				vCurItem.ParentReference.Path = vParentPath.String
			}
		}

		if vFolderChildren.Valid && vFolderChildren.Int64 > -1 {
			vCurItem.Folder = &onedriveclient.Folder{ChildCount: vFolderChildren.Int64}
		}

		if vSha1Hash.Valid {
			vCurItem.File = &onedriveclient.File{}
			vCurItem.File.Hashes.Sha1Hash = vSha1Hash.String
		}
		vRis = append(vRis, vCurItem)
	}

	return vRis, nil
}

type rawTime []byte

func (vSelf *DownTheDriveDb) GetItem(pItem interface{}, pExpandChildren bool) (vRisItem *onedriveclient.OneDriveItem, vRisError error) {

	var vConditions string
	switch pItem.(type) {

	case *onedriveclient.OneDriveItem:
		vItem := pItem.(*onedriveclient.OneDriveItem)
		vConditions = fmt.Sprintf(FIELD_ITEMS_ID+"='%s'", vItem.Id)

	case string:

		vItemString := pItem.(string)

		if vItemString == "/" {
			vConditions = FIELD_ITEMS_PARENTID + " is null"
		} else {
			if strings.Contains(vItemString, "/") {
				vConditions = fmt.Sprintf(FIELD_ITEMS_PATH+" || "+FIELD_ITEMS_NAME+" like '%s'", vItemString)
			} else {
				vConditions = fmt.Sprintf(FIELD_ITEMS_ID+"='%s'", vItemString)
			}
		}
	default:
		return nil, nil
	}

	vItems, vItemsError := vSelf.GetItemsByConditions(vConditions)

	if vItemsError != nil {
		return nil, diagnostic.NewError("error getting item %s", vItemsError, pItem)
	}

	if len(vItems) == 0 {
		return nil, nil
	}

	if len(vItems) > 1 {
		return nil, diagnostic.NewError("Invalid number of items in db that matches conditions %d", nil, vItems)
	}

	vRisItem = vItems[0]

	if vRisItem.IsFolder() && pExpandChildren {
		vChildren, vChildrenError := vSelf.GetItemsByConditions(fmt.Sprintf(FIELD_ITEMS_PARENTID+"='%s'", vRisItem.Id))

		if vChildrenError != nil {
			return nil, diagnostic.NewError("error getting children of %s", vChildrenError, pItem)
		}

		vRisItem.Children = vChildren
		//vRisItem.Folder.ChildCount = int64(len(vRisItem.Children))
	}
	return vRisItem, nil
}
