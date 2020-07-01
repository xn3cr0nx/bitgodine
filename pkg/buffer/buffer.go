package buffer

import (
	"encoding/binary"
	"errors"
)

func ReadSlice(slice *[]uint8, length uint) ([]uint8, error) {
	if len(*slice) < int(length) {
		*slice = make([]uint8, 0)
		return nil, errors.New("Eof Error")
	} else {
		res := (*slice)[:length]
		*slice = (*slice)[length:]
		return res, nil
	}
}

func ReadArray(slice *[]uint8, length uint8) (*[]uint8, error) {
	slc, err := ReadSlice(slice, uint(length))
	if err != nil {
		return nil, err
	}
	slc = slc[:length]
	return &slc, nil
}

func ReadUint8(slice *[]uint8) (uint8, error) {
	if len(*slice) == 0 {
		return 0, errors.New("Eof Error")
	} else {
		res := (*slice)[0]
		*slice = (*slice)[1:]
		return res, nil
	}
}

func ReadUint16(slice *[]uint8) (uint16, error) {
	b, err := ReadSlice(slice, 2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func ReadUint32(slice *[]uint8) (uint32, error) {
	b, err := ReadSlice(slice, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func ReadUint64(slice *[]uint8) (uint64, error) {
	b, err := ReadSlice(slice, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func ReadVarInt(slice *[]uint8) (uint64, error) {
	slice, err := ReadArray(slice, 1)
	if err != nil {
		return 0, err
	}
	n := (*slice)[0]
	switch n {
	case 0xfd:
		res, _ := ReadUint16(slice)
		return uint64(res), nil
	case 0xfe:
		res, _ := ReadUint32(slice)
		return uint64(res), nil
	case 0xff:
		res, _ := ReadUint64(slice)
		return res, nil
	default:
		return uint64(n), nil
	}
}
