package file

import (
	"io"
	"log"
	"os"
	"path/filepath"

	proto "github.com/micro/go-file/proto"
	"github.com/micro/go-micro/errors"
	"golang.org/x/net/context"
)

type handler struct {
	readDir string
	session *session
}

func (h *handler) Open(ctx context.Context, req *proto.OpenRequest, rsp *proto.OpenResponse) error {
	path := filepath.Join(h.readDir, req.Filename)
	file, err := os.Open(path)
	if err != nil {
		return errors.InternalServerError("go.micro.srv.file", err.Error())
	}

	rsp.Id = h.session.Add(file)
	rsp.Result = true

	log.Printf("Open %s, sessionId=%d", req.Filename, rsp.Id)

	return nil
}

func (h *handler) Close(ctx context.Context, req *proto.CloseRequest, rsp *proto.CloseResponse) error {
	h.session.Delete(req.Id)
	log.Printf("Close sessionId=%d", req.Id)
	return nil
}

func (h *handler) Stat(ctx context.Context, req *proto.StatRequest, rsp *proto.StatResponse) error {
	path := filepath.Join(h.readDir, req.Filename)
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.InternalServerError("go.micro.srv.file", err.Error())
	}

	if fi.IsDir() {
		rsp.Type = "Directory"
	} else {
		rsp.Type = "File"
		rsp.Size = fi.Size()
	}

	rsp.LastModified = fi.ModTime().Unix()
	log.Printf("Stat %s, %#v", req.Filename, rsp)

	return nil
}

func (h *handler) ReadAt(ctx context.Context, req *proto.ReadRequest, rsp *proto.ReadResponse) error {
	file := h.session.Get(req.Id)
	if file == nil {
		return errors.InternalServerError("go.micro.srv.file", "You must call open first.")
	}

	rsp.Data = make([]byte, req.Size)
	n, err := file.ReadAt(rsp.Data, req.Offset)
	if err != nil && err != io.EOF {
		return errors.InternalServerError("go.micro.srv.file", err.Error())
	}

	if err == io.EOF {
		rsp.Eof = true
	}

	rsp.Size = int64(n)
	rsp.Data = rsp.Data[:n]

	log.Printf("ReadAt sessionId=%d, Offset=%d, n=%d", req.Id, req.Offset, rsp.Size)

	return nil
}

func (h *handler) Read(ctx context.Context, req *proto.ReadRequest, rsp *proto.ReadResponse) error {
	file := h.session.Get(req.Id)
	if file == nil {
		return errors.InternalServerError("go.micro.srv.file", "You must call open first.")
	}

	rsp.Data = make([]byte, req.Size)
	n, err := file.Read(rsp.Data)
	if err != nil && err != io.EOF {
		return errors.InternalServerError("go.micro.srv.file", err.Error())
	}

	if err == io.EOF {
		rsp.Eof = true
	}

	rsp.Size = int64(n)
	rsp.Data = rsp.Data[:rsp.Size]

	log.Printf("Read sessionId=%d, read=%d[bytes]", req.Id, rsp.Size)

	return nil
}
