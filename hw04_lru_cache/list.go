package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length int
	front  *ListItem
	back   *ListItem
}

func NewList() List {
	return new(list)
}

func (list *list) Len() int {
	return list.length
}

func (list *list) Front() *ListItem {
	return list.front
}

func (list *list) Back() *ListItem {
	return list.back
}

func (list *list) PushFront(v interface{}) *ListItem {
	newListItem := &ListItem{
		Value: v,
		Next:  list.front,
		Prev:  nil,
	}

	if list.length == 0 {
		list.front = newListItem
		list.back = newListItem
	} else {
		list.front.Prev = newListItem
		list.front = newListItem
	}

	list.length++

	return newListItem
}

func (list *list) PushBack(v interface{}) *ListItem {
	newListItem := &ListItem{
		Value: v,
		Next:  nil,
		Prev:  list.back,
	}

	if list.length == 0 {
		list.front = newListItem
		list.back = newListItem
	} else {
		list.back.Next = newListItem
		list.back = newListItem
	}

	list.length++

	return newListItem
}

func (list *list) Remove(i *ListItem) {
	switch {
	case list.length == 1:
		list.front = nil
		list.back = nil
	case i == list.front:
		list.front = list.front.Next
		list.front.Prev = nil
	case i == list.back:
		list.back = list.back.Prev
		list.back.Next = nil
	default:
		prev := i.Prev
		next := i.Next

		prev.Next = next
		next.Prev = prev
	}

	list.length--
}

func (list *list) MoveToFront(i *ListItem) {
	switch {
	case list.front == i:
		return
	case i == list.back:
		i.Prev.Next = nil
		i.Next = list.front
		list.front.Prev = i
		list.front = i
	default:
		prev := i.Prev
		next := i.Next

		prev.Next = next
		next.Prev = prev

		i.Next = list.front
		list.front.Prev = i
		list.front = i

		i.Prev = nil
	}
}
