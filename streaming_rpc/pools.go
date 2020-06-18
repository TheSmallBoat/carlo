package streaming_rpc

import "sync"

var contextPool = &ContextPool{sp: sync.Pool{}}
