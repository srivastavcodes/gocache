package dll

import "time"

// Entry is an LRU entry.
type Entry[K comparable, V any] struct {
	// Next and prev pointers in the dll of elements. Internally, the list
	// is implemented as a ring, meaning the root element is both the next
	// element of the last list's element and the first of the previous
	// list's element.
	next, prev *Entry[K, V]

	// The list to which this element belongs.
	list *LruList[K, V]

	// The lru key of this element.
	Key K
	// The lru val of this element.
	Val V

	// The time this element would be cleaned up (optional).
	ExpiresAt time.Time

	// The expiry bucket item was put in (optional).
	ExpireBucket uint8
}

// LruList represents a doubly linked list. The zero value for LruList is an
// empty list read to use.
type LruList[K comparable, V any] struct {
	// sentinel list element, only &root, root.prev, root.next are used.
	root Entry[K, V]
	// current list length excluding sentinel element.
	len int
}

// NewList returns an initialized list.
func NewList[K comparable, V any]() *LruList[K, V] {
	return new(LruList[K, V]).Init()
}

// PrevEntry returns the previous list element or nil.
func (e *Entry[K, V]) PrevEntry() *Entry[K, V] {
	prev := e.prev
	if e.list != nil && prev != &e.list.root {
		return prev
	}
	return nil
}

// Init initializes or clears list.
func (l *LruList[K, V]) Init() *LruList[K, V] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// PushFront adds a new element to the front of list l and returns the entry.
func (l *LruList[K, V]) PushFront(k K, v V) *Entry[K, V] {
	l.lazyInit()
	return l.insertValue(k, v, time.Time{}, &l.root)
}

// PushFrontExpirable adds an expirable element to the front of list l and
// returns the entry.
func (l *LruList[K, V]) PushFrontExpirable(k K, v V, expiresAt time.Time) *Entry[K, V] {
	l.lazyInit()
	return l.insertValue(k, v, expiresAt, &l.root)
}

// Remove removes entry from its list, decrements l.len.
func (l *LruList[K, V]) Remove(entry *Entry[K, V]) V {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
	entry.next = nil // avoid memory leaks
	entry.prev = nil // avoid memory leaks
	entry.list = nil
	l.len--
	return entry.Val
}

// MoveToFront moves entry to the front of list l. It's a noOp if the entry is
// not in list l or already on the front. The element must not be nil.
func (l *LruList[K, V]) MoveToFront(entry *Entry[K, V]) {
	if entry.list != l || l.root.next == entry {
		return
	}
	l.move(entry, &l.root)
}

// Length returns the number of elements in a list. T.C. is O(1)
func (l *LruList[K, V]) Length() int { return l.len }

// Back returns the last element of the list or nil if the list is empty.
func (l *LruList[K, V]) Back() *Entry[K, V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero list value.
func (l *LruList[K, V]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts entry after index, increments l.len, and returns entry.
func (l *LruList[K, V]) insert(entry, index *Entry[K, V]) *Entry[K, V] {
	entry.prev = index
	entry.next = index.next
	entry.prev.next = entry
	entry.next.prev = entry
	entry.list = l
	l.len++
	return entry
}

// insertValue is a convenience wrapper for inserting a value with expiry
// semantics.
func (l *LruList[K, V]) insertValue(k K, v V, expiresAt time.Time, index *Entry[K, V]) *Entry[K, V] {
	entry := &Entry[K, V]{
		Key: k, Val: v,
		ExpiresAt: expiresAt,
	}
	return l.insert(entry, index)
}

// move moves entry to the next of index.
func (l *LruList[K, V]) move(entry, index *Entry[K, V]) {
	if entry == index {
		return
	}
	entry.prev.next = entry.next
	entry.next.prev = entry.prev

	entry.prev = index
	entry.next = index.next

	entry.prev.next = entry
	entry.next.prev = entry
}
