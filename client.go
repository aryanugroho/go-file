package file

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	proto "github.com/micro/go-file/proto"

	"golang.org/x/net/context"
)

const (
	blockSize = 512 * 1024
)

type fc struct {
	c proto.FileClient
}

func (c *fc) Open(filename string) (int64, error) {
	rsp, err := c.c.Open(context.TODO(), &proto.OpenRequest{Filename: filename})
	if err != nil {
		return 0, err
	}
	return rsp.Id, nil
}

func (c *fc) Stat(filename string) (*proto.StatResponse, error) {
	return c.c.Stat(context.TODO(), &proto.StatRequest{Filename: filename})
}

func (c *fc) GetBlock(sessionId, blockId int64) ([]byte, error) {
	return c.ReadAt(sessionId, blockId*blockSize, blockSize)
}

func (c *fc) ReadAt(sessionId, offset, size int64) ([]byte, error) {
	rsp, err := c.c.Read(context.TODO(), &proto.ReadRequest{Id: sessionId, Size: size, Offset: offset})
	if err != nil {
		return nil, err
	}

	if rsp.Eof {
		err = io.EOF
	}

	if rsp.Data == nil {
		rsp.Data = make([]byte, size)
	}

	if size != rsp.Size {
		return rsp.Data[:rsp.Size], err
	}

	return rsp.Data, nil
}

func (c *fc) Read(sessionId int64, buf []byte) (int, error) {
	b, err := c.ReadAt(sessionId, 0, int64(cap(buf)))
	if err != nil {
		return 0, err
	}
	copy(buf, b)
	return len(b), nil
}

func (c *fc) Close(sessionId int64) error {
	_, err := c.c.Close(context.TODO(), &proto.CloseRequest{Id: sessionId})
	return err
}

func (c *fc) Download(filename, saveFile string) error {
	return c.DownloadAt(filename, saveFile, 0)
}

func (c *fc) DownloadAt(filename, saveFile string, blockId int) error {
	stat, err := c.Stat(filename)
	if err != nil {
		return err
	}
	if stat.Type == "Directory" {
		return errors.New(fmt.Sprintf("%s is directory.", filename))
	}

	blocks := int(stat.Size / blockSize)
	if stat.Size%blockSize != 0 {
		blocks += 1
	}

	log.Printf("Download %s in %d blocks\n", filename, blocks-blockId)

	file, err := os.OpenFile(saveFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	sessionId, err := c.Open(filename)
	if err != nil {
		return err
	}

	for i := blockId; i < blocks; i++ {
		buf, rerr := c.GetBlock(sessionId, int64(i))
		if rerr != nil && rerr != io.EOF {
			return rerr
		}
		if _, werr := file.WriteAt(buf, int64(i)*blockSize); werr != nil {
			return werr
		}

		if i%((blocks-blockId)/100+1) == 0 {
			log.Printf("Downloading %s [%d/%d] blocks", filename, i-blockId+1, blocks-blockId)
		}

		if rerr == io.EOF {
			break
		}
	}
	log.Printf("Download %s completed", filename)

	c.Close(sessionId)

	return nil
}
