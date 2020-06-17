package lib

type Context struct {
	conn *Conn
	seq  uint32
	buf  []byte
}

func (c *Context) Conn() *Conn            { return c.conn }
func (c *Context) Body() []byte           { return c.buf }
func (c *Context) Reply(buf []byte) error { return c.conn.send(c.seq, buf) }
