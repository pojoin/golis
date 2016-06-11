package golis

type IoFilter struct {
	name string
	prev *IoFilter
	next *IoFilter
}

func (i *IoFilter) Next() *IoFilter {
	return i.next
}

func (i *IoFilter) SessionOpened(session Iosession) {
	i.Next().SessionOpened(session)
}

func (i *IoFilter) SessionClosed(session Iosession) {
	i.Next().SessionClosed(session)
}

func (i *IoFilter) MessageReceived(session Iosession, message interface{}) {
	i.Next().MessageReceived(session, message)
}

func (i *IoFilter) MessageSent(session Iosession, message interface{}) {
	i.Next().MessageSent(session, message)
}
func (i *IoFilter) ErrorCaught(session Iosession, err error) {
	i.Next().ErrorCaught(session, err)
}

type IoFilterChain struct {
	head *IoFilter
}

func (f *IoFilterChain) AddLast(name string, filter *IoFilter) *IoFilterChain {
	if f.head == nil {
		filter.next = nil
		filter.prev = nil
		f.head = filter
	} else {
		lastFilter := getLastFilter(f.head)
		lastFilter.next = &IoFilter{
			name: name,
			next: nil,
			prev: lastFilter,
		}
	}
	return f
}

func (f *IoFilterChain) AddAfter(baseName, name string, filter *IoFilter) *IoFilterChain {
	if e, ok := getFilterByName(baseName, f.head); ok {
		tmp := e.next
		e.next = &IoFilter{
			name: name,
			next: tmp,
			prev: e,
		}
		tmp.prev = e.next
	}
	return f
}

func (f *IoFilterChain) AddBefore(baseName, name string, filter *IoFilter) *IoFilterChain {
	if e, ok := getFilterByName(baseName, f.head); ok {
		tmp := e.prev
		e.prev = &IoFilter{
			name: name,
			next: e,
			prev: tmp,
		}
		tmp.next = e.prev
	}
	return f
}

func getFilterByName(name string, root *IoFilter) (*IoFilter, bool) {
	if root == nil {
		return nil, false
	}
	if root.name == name {
		return root, true
	}
	return getFilterByName(name, root.next)
}

func getLastFilter(head *IoFilter) *IoFilter {
	if head.next == nil {
		return head
	}
	return getLastFilter(head.next)
}
