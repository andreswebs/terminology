package tbx_test

import (
	"errors"
	"testing"

	"github.com/andreswebs/terminology/internal/tbx"
)

func TestCheckDoctype_BareAccepted(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><!DOCTYPE tbx><tbx></tbx>`)
	if err := tbx.CheckDoctype(data); err != nil {
		t.Errorf("bare DOCTYPE should be accepted, got: %v", err)
	}
}

func TestCheckDoctype_NoDoctype(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><tbx></tbx>`)
	if err := tbx.CheckDoctype(data); err != nil {
		t.Errorf("no DOCTYPE should be accepted, got: %v", err)
	}
}

func TestCheckDoctype_InternalSubsetRejected(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><!DOCTYPE tbx [<!ENTITY xxe "boom">]><tbx></tbx>`)
	err := tbx.CheckDoctype(data)
	if err == nil {
		t.Fatal("DOCTYPE with internal subset should be rejected")
	}
	var coded interface{ Code() string }
	if !errors.As(err, &coded) || coded.Code() != "invalid_input" {
		t.Errorf("expected code invalid_input, got: %v", err)
	}
}

func TestCheckDoctype_SystemRejected(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><!DOCTYPE tbx SYSTEM "http://evil.com/xxe.dtd"><tbx></tbx>`)
	err := tbx.CheckDoctype(data)
	if err == nil {
		t.Fatal("DOCTYPE with SYSTEM should be rejected")
	}
	var coded interface{ Code() string }
	if !errors.As(err, &coded) || coded.Code() != "invalid_input" {
		t.Errorf("expected code invalid_input, got: %v", err)
	}
}

func TestCheckDoctype_PublicRejected(t *testing.T) {
	data := []byte(`<?xml version="1.0"?><!DOCTYPE tbx PUBLIC "-//TBX//DTD" "http://example.com/tbx.dtd"><tbx></tbx>`)
	err := tbx.CheckDoctype(data)
	if err == nil {
		t.Fatal("DOCTYPE with PUBLIC should be rejected")
	}
	var coded interface{ Code() string }
	if !errors.As(err, &coded) || coded.Code() != "invalid_input" {
		t.Errorf("expected code invalid_input, got: %v", err)
	}
}
