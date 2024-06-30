package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // 底层连接，实现了 ReadWriteCloser 接口，提供读写和关闭功能
	buf  *bufio.Writer      // 带缓冲的写入器，用于优化写入性能
	dec  *gob.Decoder       // Gob 解码器，用于解码从连接中读取的数据
	enc  *gob.Encoder       // Gob 编码器，用于将数据编码并写入缓冲区
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (g *GobCodec) ReadHeader(header *Header) error {
	// Decode will block if not have input;
	// if read eof it will return io.EOF ERR
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		log.Println("rpc: gob error encoding header:", err)
		return
	}

	if err = g.enc.Encode(body); err != nil {
		log.Println("rpc: gob error encoding body:", err)
		return
	}
	return
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}
