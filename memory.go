package session_memory

import (
	"container/list"
	"sync"
	"time"

	session "github.com/snowuly/session-go"
)

var pdr = &Provider{list: list.New()}

func init() {
	pdr.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", pdr)
}

type Session struct {
	accessTime time.Time
	store      map[string]interface{}
	sid        string
}

func (s *Session) Get(key string) interface{} {
	pdr.sessionUpdate(s.sid)
	return s.store[key]
}
func (s *Session) Set(key string, value interface{}) {
	s.store[key] = value
	pdr.sessionUpdate(s.sid)
}
func (s *Session) Del(key string) {
	delete(s.store, key)
	pdr.sessionUpdate(s.sid)
}
func (s *Session) SessionId() string {
	return s.sid
}

type Provider struct {
	lock     sync.RWMutex
	sessions map[string]*list.Element
	list     *list.List
}

func (p *Provider) SessionInit(sid string) (session session.Session, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	session = &Session{time.Now(), make(map[string]interface{}), sid}
	element := p.list.PushFront(session)
	p.sessions[sid] = element
	return
}

func (p *Provider) SessionRead(sid string) (session.Session, error) {
	if element, ok := p.sessions[sid]; ok {
		return element.Value.(*Session), nil
	}
	return p.SessionInit(sid)
}
func (p *Provider) SessionDestroy(sid string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if element, ok := p.sessions[sid]; ok {
		delete(p.sessions, sid)
		p.list.Remove(element)
	}
}

func (p *Provider) sessionUpdate(sid string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	element, ok := p.sessions[sid]
	if !ok {
		return
	}
	session := element.Value.(*Session)
	session.accessTime = time.Now()
	p.list.MoveToFront(element)
}
func (p *Provider) GC(maxLifetime int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for {
		element := p.list.Back()
		if element == nil {
			break
		}
		session := element.Value.(*Session)
		if session.accessTime.Unix()+int64(maxLifetime) < time.Now().Unix() {
			delete(p.sessions, session.sid)
			p.list.Remove(element)
		} else {
			break
		}
	}
}
