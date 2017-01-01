package fs

import (
	"bazil.org/fuse"
	"github.com/mysinmyc/gocommons/diagnostic"
	"golang.org/x/net/context"
	"io"
	"os"
)

type FileHandle struct {
	file   *os.File
}


func (vSelf *FileHandle) Read(pContext context.Context, pRequest *fuse.ReadRequest, pResponse *fuse.ReadResponse) error {
	vBytes := make([]byte, pRequest.Size)
	vBytesRead, vReadError := vSelf.file.ReadAt(vBytes, pRequest.Offset)
	if vReadError != nil && vReadError != io.EOF {
		return diagnostic.NewError("Failed to read file", vReadError)
	}
	pResponse.Data = vBytes[:vBytesRead]
	return nil
}

func (vSelf *FileHandle) Release(pContext context.Context, pRequest *fuse.ReleaseRequest) error {
	return vSelf.file.Close()
}
