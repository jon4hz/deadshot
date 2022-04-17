package database

import (
	"errors"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var ErrNoEndpointsFound = errors.New("no endpoints found")

// Endpoint represents a node endpoint, either http or ws.
type Endpoint struct {
	gorm.Model `yaml:"-"`
	URL        string     `yaml:"url"`
	Custom     bool       `yaml:"-"`
	NetworkID  uint       `yaml:"-"`
	mu         sync.Mutex `yaml:"-" gorm:"-"`
}

// NewEndpoint creates a new endpoint.
func NewEndpoint(url string, custom bool) *Endpoint {
	return &Endpoint{
		URL:    url,
		Custom: custom,
	}
}

// Get returns the endpoint.
func (e *Endpoint) Get() *Endpoint {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e
}

// GetURL returns the url of the endpoint.
func (e *Endpoint) GetURL() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.URL
}

// SetUrl sets the url of the endpoint.
func (e *Endpoint) SetURL(url string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.URL = url
}

func (e *Endpoint) GetCustom() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Custom
}

type Endpoints []*Endpoint

// GetEndpointByURL returns the endpoint with the given url.
func (e Endpoints) GetEndpointByURL(url string) *Endpoint {
	for _, endpoint := range e {
		if strings.EqualFold(endpoint.URL, url) {
			return endpoint
		}
	}
	return nil
}

// GetUrls returns the urls of the endpoints.
func (e Endpoints) GetUrls() []string {
	urls := make([]string, len(e))
	for i, endpoint := range e {
		urls[i] = endpoint.URL
	}
	return urls
}
