package files

import (
	"github.com/status-im/status-go/files"
	"github.com/status-im/status-go/protocol/protobuf"
)

func FileType(buf []byte) protobuf.FileType {
	switch files.GetType(buf) {
	case files.PDF:
		return protobuf.FileType_PDF
	case files.DOC:
		return protobuf.FileType_DOC
	case files.XLS:
		return protobuf.FileType_XLS
	case files.WEBP:
		return protobuf.FileType_WEBP
	default:
		return protobuf.FileType_UNKNOWN_File_TYPE
	}
}