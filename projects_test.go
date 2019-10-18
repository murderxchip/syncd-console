package main

import "testing"

func TestGet(t *testing.T){
	p := NewProjects()
	t.Log(p.GetProject("test-admin-server"))
}
