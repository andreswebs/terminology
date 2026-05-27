package tbx

import (
	"errors"
	"io"
	"testing"
)

type stubReader struct{}

func (stubReader) Decode(io.Reader) (*Glossary, []Warning, error) { return nil, nil, nil }

type stubWriter struct{}

func (stubWriter) Encode(io.Writer, *Glossary) error { return nil }

func TestReaderFor_Registered(t *testing.T) {
	const testDialect Dialect = "test-dialect-reader"
	RegisterDialect(testDialect, func() Reader { return stubReader{} }, func() Writer { return stubWriter{} })
	t.Cleanup(func() { unregisterDialect(testDialect) })

	r, err := readerFor(testDialect)
	if err != nil {
		t.Fatalf("readerFor: %v", err)
	}
	if r == nil {
		t.Fatal("readerFor returned nil")
	}
}

func TestReaderFor_Unregistered(t *testing.T) {
	_, err := readerFor("no-such-dialect")
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("got %v, want ErrUnsupportedDialect", err)
	}
}

func TestWriterFor_Registered(t *testing.T) {
	const testDialect Dialect = "test-dialect-writer"
	RegisterDialect(testDialect, func() Reader { return stubReader{} }, func() Writer { return stubWriter{} })
	t.Cleanup(func() { unregisterDialect(testDialect) })

	w, err := writerFor(testDialect)
	if err != nil {
		t.Fatalf("writerFor: %v", err)
	}
	if w == nil {
		t.Fatal("writerFor returned nil")
	}
}

func TestWriterFor_Unregistered(t *testing.T) {
	_, err := writerFor("no-such-dialect")
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("got %v, want ErrUnsupportedDialect", err)
	}
}

func TestRegisterDialect_FactoryFreshInstances(t *testing.T) {
	const testDialect Dialect = "test-dialect-fresh"
	calls := 0
	RegisterDialect(testDialect, func() Reader {
		calls++
		return stubReader{}
	}, func() Writer { return stubWriter{} })
	t.Cleanup(func() { unregisterDialect(testDialect) })

	_, _ = readerFor(testDialect)
	_, _ = readerFor(testDialect)
	if calls != 2 {
		t.Errorf("factory called %d times, want 2", calls)
	}
}
