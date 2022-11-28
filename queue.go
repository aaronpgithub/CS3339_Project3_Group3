package main

// queue dtype holds a dynamically sized slice of any type of data.
// Queue is First in, First out. (FIFO)
type Queue struct {
	data    []interface{}
	maxSize int
}

func initQueue(maxSize int) Queue {
	var temp = Queue{}
	temp.maxSize = maxSize

	return temp
}

// append new value to queue if not full
//
//	returns queue and error value, 0 if ok, 1 if cannot insert data because queue is full.
func (q Queue) enqueue(data interface{}) (Queue, int) {
	var err = 0

	if q.maxSize > len(q.data) {
		q.data = append(q.data, data)
	} else {
		err = 1
	}

	return q, err
}

// returns queue, dequeued value and error value, 0 if ok, 1 if cannot delete data
func (q Queue) dequeue() (Queue, interface{}, int) {
	var err = 0
	var value = q.data[0]

	q.data = q.data[1:] // Slice off the element once it is dequeued.
	return q, value, err
}
