package dotnet

import (
    "context"
    "net"
    "net/http"
    "sync"
)

type clientPool struct {
    pool sync.Pool
}

func newClientPool(socket string) *clientPool {
    return &clientPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &http.Client{
                    Transport: &http.Transport{
                        DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
                            return net.Dial("unix", socket)
                        },
                    },
                }
            },
        },
    }
}

func (cp *clientPool) get() *http.Client {
    return cp.pool.Get().(*http.Client)
}

func (cp *clientPool) put(client *http.Client) {
    cp.pool.Put(client)
}