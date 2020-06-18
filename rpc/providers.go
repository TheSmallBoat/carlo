package rpc

import (
	"sync"

	carlo "github.com/TheSmallBoat/carlo/lib"
	"github.com/lithdew/kademlia"
)

type Providers struct {
	sync.Mutex

	services  map[string]map[*carlo.Conn]struct{}
	providers map[*carlo.Conn]*Provider
}

func NewProviders() *Providers {
	return &Providers{
		services:  make(map[string]map[*carlo.Conn]struct{}),
		providers: make(map[*carlo.Conn]*Provider),
	}
}

func (p *Providers) findProvider(conn *carlo.Conn) *Provider {
	p.Lock()
	defer p.Unlock()
	return p.providers[conn]
}

func (p *Providers) getProviders(services ...string) []*Provider {
	p.Lock()
	defer p.Unlock()

	var conns []*carlo.Conn

	for _, service := range services {
		for conn := range p.services[service] {
			conns = append(conns, conn)
		}
	}

	if conns == nil {
		return nil
	}

	providers := make([]*Provider, 0, len(conns))
	for _, conn := range conns {
		providers = append(providers, p.providers[conn])
	}

	return providers
}

func (p *Providers) registerProvider(conn *carlo.Conn, id *kademlia.ID, services []string, outgoing bool) (*Provider, bool) {
	p.Lock()
	defer p.Unlock()

	provider, exists := p.providers[conn]
	if !exists {
		provider = &Provider{
			services: make(map[string]struct{}),
			streams:  make(map[uint32]*Stream),
		}
		if outgoing {
			provider.counter = 1
		} else {
			provider.counter = 0
		}
		p.providers[conn] = provider
	}

	provider.kadId = id
	provider.conn = conn

	for _, service := range services {
		provider.services[service] = struct{}{}
		if _, exists := p.services[service]; !exists {
			p.services[service] = make(map[*carlo.Conn]struct{})
		}
		p.services[service][conn] = struct{}{}
	}

	return provider, exists
}

func (p *Providers) deregisterProvider(conn *carlo.Conn) *Provider {
	p.Lock()
	defer p.Unlock()

	provider, exists := p.providers[conn]
	if !exists {
		return nil
	}

	delete(p.providers, conn)

	for service := range provider.services {
		delete(p.services[service], conn)
		if len(p.services[service]) == 0 {
			delete(p.services, service)
		}
	}

	provider.Close()

	return provider
}
