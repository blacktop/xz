package xz

import (
	"errors"
	"io"

	"github.com/ulikunitz/xz/lzma2"
)

// LZMA filter constants.
const (
	lzmaFilterID  = 0x21
	lzmaFilterLen = 3
)

// lzmaFilter declares the LZMA2 filter information stored in an xz
// block header.
type lzmaFilter struct {
	dictCap int64
}

// id returns the ID for the LZMA2 filter.
func (f lzmaFilter) id() uint64 { return lzmaFilterID }

// MarshalBinary converts the lzmaFilter in its encoded representation.
func (f lzmaFilter) MarshalBinary() (data []byte, err error) {
	c := lzma2.EncodeDictCap(f.dictCap)
	return []byte{lzmaFilterID, 1, c}, nil
}

// UnmarshalBinary unmarshals the given data representation of the LZMA2
// filter.
func (f *lzmaFilter) UnmarshalBinary(data []byte) error {
	if len(data) != lzmaFilterLen {
		return errors.New("xz: data for LZMA2 filter has wrong length")
	}
	if data[0] != lzmaFilterID {
		return errors.New("xz: wrong LZMA2 filter id")
	}
	if data[1] != 1 {
		return errors.New("xz: wrong LZMA2 filter size")
	}
	dc, err := lzma2.DecodeDictCap(data[2])
	if err != nil {
		return errors.New("xz: wrong LZMA2 dictionary size property")
	}

	f.dictCap = dc
	return nil
}

// reader creates a new reader for the LZMA2 filter.
func (f lzmaFilter) reader(r io.Reader, p *ReaderParams) (fr io.Reader,
	err error) {

	params := p.ReaderParams
	dc := int(f.dictCap)
	if dc < 1 {
		return nil, errors.New("xz: LZMA2 filter parameter " +
			"dictionary capacity overflow")
	}
	if dc > params.DictCap {
		params.DictCap = dc
	}

	fr, err = lzma2.NewReaderParams(r, &params)
	if err != nil {
		return nil, err
	}
	return fr, nil
}

// writeCloser creates a io.WriteCloser for the LZMA2 filter.
func (f lzmaFilter) writeCloser(w io.WriteCloser, p *WriterParams,
) (fw io.WriteCloser, err error) {

	params := p.WriterParams
	dc := int(f.dictCap)
	if dc < 1 {
		return nil, errors.New("xz: LZMA2 filter parameter " +
			"dictionary capacity overflow")
	}
	if dc > params.DictCap {
		params.DictCap = dc
	}

	fw, err = lzma2.NewWriterParams(w, &params)
	if err != nil {
		return nil, err
	}
	return fw, nil
}

// last returns true, because an LZMA2 filter must be the last filter in
// the filter list.
func (f lzmaFilter) last() bool { return true }
