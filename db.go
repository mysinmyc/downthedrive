package downthedrive

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mysinmyc/gocommons/db"
	"github.com/mysinmyc/gocommons/diagnostic"
)

type DownTheDriveDb struct {
	db.DbHelper
	itemBulk   bool
	itemInsert *db.SqlInsert
}

func NewDb(pDriver string, pDataSourceName string) (vRisIndexDb *DownTheDriveDb, vError error) {

	vRis := &DownTheDriveDb{}
	vInitError := vRis.initDb(pDriver, pDataSourceName)

	if vInitError != nil {
		return nil, diagnostic.NewError("Failed to initialize db ", vInitError)
	}
	return vRis, nil
}

func (vSelf *DownTheDriveDb) initDb(pDriver string, pDataSourceName string) error {

	vDbHelper, vError := db.NewDbHelper(pDriver, pDataSourceName)
	if vError != nil {
		return diagnostic.NewError("Error opening db", vError)
	}
	vSelf.DbHelper = *vDbHelper

	switch vDbHelper.GetDbType() {
	case db.DbType_sqlite3:
		vSelf.DbHelper.SetMaxOpenConns(1)
		_, vError = vSelf.DbHelper.Exec(DDL_ITEMS_SQLITE)
	case db.DbType_mysql:
		_, vError = vSelf.DbHelper.Exec(DDL_ITEMS_MYSQL)
	default:
		return diagnostic.NewError("Unsupported db type %s", nil, vDbHelper.GetDbType())
	}
	if vError != nil {
		return diagnostic.NewError("failed to create items table", vError)
	}

	return nil
}

