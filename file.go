package file

import (
	"os"

	proto "github.com/micro/go-file/proto"
	"github.com/micro/go-micro/client"
)

// Client is the client interface to access files
type Client interface {
	Open(filename string) (int64, error)
	Stat(filename string) (*proto.StatResponse, error)
	GetBlock(sessionId, blockId int64) ([]byte, error)
	ReadAt(sessionId, offset, size int64) ([]byte, error)
	Read(sessionId int64, buf []byte) (int, error)
	Close(sessionId int64) error
	Download(filename, saveFile string) error
	DownloadAt(filename, saveFile string, blockId int) error
}

// NewClient returns a new Client which uses a micro Client
func NewClient(service string, c client.Client) Client {
	return &fc{proto.NewFileClient(service, c)}
}

// NewHandler is a handler that can be registered with a micro Server
func NewHandler(readDir string) proto.FileHandler {
	return &handler{
		readDir: readDir,
		session: &session{
			files: make(map[int64]*os.File),
		},
	}
}
