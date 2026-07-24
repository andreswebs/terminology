package tbx

var (
	readers = map[Dialect]func() Reader{}
	writers = map[Dialect]func() Writer{}
)

// RegisterDialect registers the reader and writer constructors for dialect d.
func RegisterDialect(d Dialect, rf func() Reader, wf func() Writer) {
	readers[d] = rf
	writers[d] = wf
}

func readerFor(d Dialect) (Reader, error) {
	rf, ok := readers[d]
	if !ok {
		return nil, ErrUnsupportedDialect
	}
	return rf(), nil
}

// ReaderForDialect returns a Reader for dialect d, or ErrUnsupportedDialect if
// none is registered.
func ReaderForDialect(d Dialect) (Reader, error) {
	return readerFor(d)
}

// WriterForDialect returns a Writer for dialect d, or ErrUnsupportedDialect if
// none is registered.
func WriterForDialect(d Dialect) (Writer, error) {
	return writerFor(d)
}

func writerFor(d Dialect) (Writer, error) {
	wf, ok := writers[d]
	if !ok {
		return nil, ErrUnsupportedDialect
	}
	return wf(), nil
}

func unregisterDialect(d Dialect) {
	delete(readers, d)
	delete(writers, d)
}
