package main

import "testing"

func TestReadFromClosed(t *testing.T) {
	c := make(chan int)
	close(c)

	res := <-c
	t.Logf("read from closed: %v", res)
}

func TestWriteToClosed(t *testing.T) {
	c := make(chan int)
	close(c)

	c <- 42
}

func TestCloseClosed(t *testing.T) {
	c := make(chan int)
	close(c)
	close(c)
}

func TestReadFromNil(t *testing.T) {
	var c chan int

	res, ok := <-c
	t.Logf("read from nil: %v %v", res, ok)
}

func TestWriteToNil(t *testing.T) {
	var c chan int

	c <- 42
}

func TestCloseNil(t *testing.T) {
	var c chan int

	close(c)
}
