package golis

import "errors"

type IoFilter interface {
	SessionOpened(session *Iosession) bool
	SessionClosed(session *Iosession) bool
	MsgReceived(session *Iosession, message interface{}) bool
	MsgSend(session *Iosession, message interface{}) bool
	ErrorCaught(sesion *Iosession, err error) bool
}

type IoFilterAdapter struct{}

func (*IoFilterAdapter) SessionOpened(session *Iosession) bool                    { return true }
func (*IoFilterAdapter) SessionClosed(session *Iosession) bool                    { return true }
func (*IoFilterAdapter) ErrorCaught(session *Iosession, err error) bool           { return true }
func (*IoFilterAdapter) MsgSend(session *Iosession, message interface{}) bool     { return true }
func (*IoFilterAdapter) MsgReceived(session *Iosession, message interface{}) bool { return true }

type Codecer interface {
	Decode(buffer *Buffer, dataCh chan<- interface{}) error
	Encode(message interface{}) ([]byte, error)
}

type ProtocalCodec struct {
}

func (*ProtocalCodec) Decode(buffer *Buffer, dataCh chan<- interface{}) error {
	return nil
}
func (*ProtocalCodec) Encode(message interface{}) ([]byte, error) {
	if bs, ok := message.([]byte); ok {
		return bs, nil
	}
	return nil, errors.New("codec failed")
}

type entry struct {
	name   string
	filter IoFilter
	prev   *entry
	next   *entry
}

func (e *entry) sessionOpened(session *Iosession) bool {
	if e.filter.SessionOpened(session) {
		if e.next != nil {
			return e.next.sessionOpened(session)
		} else {
			return true
		}
	}
	return false
}

func (e *entry) sessionClosed(session *Iosession) bool {
	if e.filter.SessionClosed(session) {
		if e.next != nil {
			return e.next.sessionClosed(session)
		} else {
			return true
		}
	}
	return false
}

func (e *entry) msgReceived(session *Iosession, message interface{}) bool {
	var flag bool
	if flag = e.filter.MsgReceived(session, message); flag && e.next != nil {
		return e.next.msgReceived(session, message)
	}
	return flag
}

func (e *entry) errorCaught(session *Iosession, err error) bool {
	if e.filter.ErrorCaught(session, err) {
		if e.next != nil {
			return e.next.errorCaught(session, err)
		} else {
			return true
		}
	}
	return false
}
func (e *entry) msgSend(session *Iosession, message interface{}) (interface{}, bool) {
	var flag bool
	if flag = e.filter.MsgSend(session, message); flag && e.prev != nil {
		return e.prev.msgSend(session, message)
	}
	return message, flag
}

// when invok return true ,it will next
type IoFilterChain struct {
	head *entry
}

func (f *IoFilterChain) sessionOpened(session *Iosession) bool {
	return f.head.sessionOpened(session)
}

func (f *IoFilterChain) sessionClosed(session *Iosession) bool {
	return f.head.sessionClosed(session)
}

func (f *IoFilterChain) msgReceived(session *Iosession, message interface{}) bool {
	return f.head.msgReceived(session, message)
}

func (f *IoFilterChain) errorCaught(session *Iosession, err error) bool {
	return f.head.errorCaught(session, err)
}

func (f *IoFilterChain) msgSend(session *Iosession, message interface{}) (interface{}, bool) {
	lastEntry := getLastEntry(f.head)
	return lastEntry.msgSend(session, message)
}

func (f *IoFilterChain) AddLast(name string, filter IoFilter) *IoFilterChain {
	if f.head == nil {
		f.head = &entry{
			name:   name,
			prev:   nil,
			next:   nil,
			filter: filter,
		}
	} else {
		lastEntry := getLastEntry(f.head)
		lastEntry.next = &entry{
			name:   name,
			filter: filter,
			next:   nil,
			prev:   lastEntry,
		}
	}
	return f
}

func (f *IoFilterChain) AddAfter(baseName, name string, filter IoFilter) *IoFilterChain {
	if e, ok := getEntryByName(baseName, f.head); ok {
		tmp := e.next
		e.next = &entry{
			name:   name,
			filter: filter,
			next:   tmp,
			prev:   e,
		}
		if tmp != nil {
			tmp.prev = e.next
		}
	}
	return f
}

func (f *IoFilterChain) AddBefore(baseName, name string, filter IoFilter) *IoFilterChain {
	if e, ok := getEntryByName(baseName, f.head); ok {
		tmp := e.prev
		e.prev = &entry{
			name:   name,
			filter: filter,
			next:   e,
			prev:   tmp,
		}
		if tmp != nil {
			tmp.next = e.prev
		}
	}
	return f
}

func getEntryByName(name string, root *entry) (*entry, bool) {
	if root == nil {
		return nil, false
	}
	if root.name == name {
		return root, true
	}
	return getEntryByName(name, root.next)
}

func getLastEntry(head *entry) *entry {
	if head.next == nil {
		return head
	}
	return getLastEntry(head.next)
}
