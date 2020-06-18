package rpc

import "sync"

var contextPool = &ContextPool{sp: sync.Pool{}}
