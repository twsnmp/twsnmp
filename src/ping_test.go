package main

import (
	"context"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	pctx := context.Background()
	ctx, cancel := context.WithCancel(pctx)
	go pingBackend(ctx)
	defer cancel()
	time.Sleep(time.Second * 1)
	r := doPing("192.168.1.1", 1, 1, 12)
	if r.Stat != pingOK {
		t.Errorf("ping stat = %d", r.Stat)
	}
	t.Log("Done")
}
