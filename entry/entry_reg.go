package rk

import (
	"context"
	"embed"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type entryByName map[string]Entry

type entryByKind map[string]entryByName

type embedFsByName map[string]*embed.FS

type embedFsByKind map[string]embedFsByName

type ShutdownHook func()

var (
	eventIdRaw, _ = uuid.NewUUID()

	Registry = &registry{
		mu:            sync.Mutex{},
		entries:       make(entryByKind),
		embedFS:       make(embedFsByKind),
		shutdownHooks: make([]ShutdownHook, 0),
		shutdownOnce:  sync.Once{},
		waitOnce:      sync.Once{},
		startTime:     time.Now(),
		eventId:       eventIdRaw.String(),
	}
)

func init() {
	signal.Notify(Registry.shutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	Registry.raw, Registry.shutdownFn = context.WithCancel(context.Background())
}

type registry struct {
	mu  sync.Mutex
	raw context.Context

	shutdownFn    context.CancelFunc
	shutdownSig   chan os.Signal
	shutdownErr   error
	shutdownHooks []ShutdownHook
	shutdownOnce  sync.Once

	waitOnce sync.Once

	entries entryByKind
	embedFS embedFsByKind

	values map[interface{}]interface{}

	startTime      time.Time
	serviceName    string
	serviceVersion string
	eventId        string
}

func (c *registry) SetDeadline(deadline time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.raw, _ = context.WithDeadline(c.raw, deadline)
}

func (c *registry) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.shutdownFn()
	})
}

func (c *registry) ShutdownErr() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.shutdownErr
}

func (c *registry) Wait() {
	c.waitOnce.Do(func() {
		select {
		case <-c.raw.Done():
			c.mu.Lock()
			c.mu.Unlock()
			if err := c.raw.Err(); err != nil {
				c.shutdownErr = err
			} else {
				c.shutdownErr = fmt.Errorf("unknown reason")
			}
		case v := <-c.shutdownSig:
			c.mu.Lock()
			c.mu.Unlock()
			c.shutdownErr = fmt.Errorf("shutdown by signal %s", v.String())
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		for i := range c.shutdownHooks {
			c.shutdownHooks[i]()
		}
	})
}

// ****** Value related ******

// AddValue add value to global context.
func (c *registry) AddValue(key interface{}, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// GetValue returns value from global context.
func (c *registry) GetValue(key interface{}) interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.values[key]
}

// ListValues list values from global context.
func (c *registry) ListValues() map[interface{}]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.values
}

// RemoveValue remove value from global context.
func (c *registry) RemoveValue(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

// ClearValues clear values from global context.
func (c *registry) ClearValues() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.values {
		delete(c.values, k)
	}
}

// ****** Entry related ******

func (c *registry) AddEntry(e Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, contains := c.entries[e.Kind()]; !contains {
		c.entries[e.Kind()] = make(entryByName)
	}

	c.entries[e.Kind()][e.Name()] = e
}

func (c *registry) RemoveEntry(kind, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if byName, contains := c.entries[kind]; contains {
		delete(byName, name)
	}
}

// ListEntriesByKind list entries by kind
func (c *registry) ListEntriesByKind(kind string) []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := make([]Entry, 0)

	if byName, ok := c.entries[kind]; ok {
		for _, e := range byName {
			res = append(res, e)
		}
	}

	return res
}

// GetEntry get Entry by kind and name
func (c *registry) GetEntry(kind, name string) Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if byName, ok := c.entries[kind]; ok {
		if v, ok := byName[name]; ok {
			return v
		}
	}
	return nil
}

func (c *registry) GetEntryOrDefault(kind, name string) Entry {
	c.mu.Lock()
	defer c.mu.Unlock()

	if byName, ok := c.entries[kind]; ok {
		if v, ok := byName[name]; ok {
			return v
		}
	}

	// search for default
	if byName, ok := c.entries[kind]; ok {
		for _, v := range byName {
			if v.Config().Header().Metadata.Default {
				return v
			}
		}
	}

	return nil
}

// ****** EntryFs related ******

// MapEntryFS add embed.FS based on name and type of Entry
func (c *registry) MapEntryFS(kind, name string, fs *embed.FS) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(kind) < 1 || len(kind) < 1 || fs == nil {
		return
	}

	if _, ok := c.embedFS[kind]; !ok {
		c.embedFS[kind] = make(embedFsByName)
	}

	c.embedFS[kind][name] = fs
}

func (c *registry) EntryFS(kind, name string) *embed.FS {
	c.mu.Lock()
	defer c.mu.Unlock()

	if byName, ok := c.embedFS[kind]; ok {
		return byName[name]
	}

	return nil
}

func (c *registry) ServiceName() string {
	return c.serviceName
}

func (c *registry) ServiceVersion() string {
	return c.serviceVersion
}

func (c *registry) EventId() string {
	return c.eventId
}
