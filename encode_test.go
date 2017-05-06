package d2b

import (
	"encoding/binary"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEncode(t *testing.T) {
	Convey("Test Encode", t, func() {
		type CustomType [5][2]int16
		type TestStruct1 struct {
			FieldA string  `d2b:"length:10"`
			FieldB string  `d2b:"length:10"`
			Slice  []int32 `d2b:"length:2"`
			A      CustomType
			QQ     *int32
			B      int16
			C      int32
			D      int64
			E      uint8
			F      uint16
			G      uint32
			H      uint64
			EE     string `d2b:"-"`
			I      int8
			Q      *uint8
		}
		Convey("Should return empty byte array if nil passed", func() {
			type a **TestStruct1
			var data a
			bytes, err := Encode(data, binary.LittleEndian)
			if err != nil {
				t.Error(err)
				return
			}
			So(err, ShouldBeNil)
			So(bytes, ShouldHaveLength, 83)
		})
		Convey("Should return empty byte array if nil array passed", func() {
			var data [5]int32
			bytes, err := Encode(data, binary.LittleEndian)
			if err != nil {
				t.Error(err)
				return
			}
			So(err, ShouldBeNil)
			So(bytes, ShouldHaveLength, 20)
		})
		Convey("Should return valid byte array", func() {
			data := &TestStruct1{
				FieldA: "Hello",
				FieldB: "World",
				QQ:     new(int32),
				Slice:  []int32{1},
			}
			bytes, err := Encode(data, binary.LittleEndian)
			So(err, ShouldBeNil)
			So(bytes, ShouldHaveLength, 83)
		})
		Convey("Should return error if struct tag length contains wrong value", func() {
			type ErrTestStruct struct {
				Field string `d2b:"length:1qwe"`
			}
			bytes, err := Encode(&ErrTestStruct{Field: "Hello"}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})

		Convey("Should return error if struct contains slice field with type, that's not valid for marshalling", func() {
			type ErrTestStruct struct {
				Field []int `d2b:"length:2"` //31
			}
			bytes, err := Encode(&ErrTestStruct{Field: []int{10}}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if struct contains slice field with type, that's not valid for marshalling 2", func() {
			type ErrTestStruct struct {
				Field []int `d2b:"length:2"` //31
			}
			bytes, err := Encode(&ErrTestStruct{}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if struct contains array field with type, that's not valid for marshalling", func() {
			type ErrTestStruct struct {
				Field [2]int
			}
			bytes, err := Encode(&ErrTestStruct{Field: [2]int{10, 11}}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if struct slice field doesn't contains length param", func() {
			type ErrTestStruct struct {
				Field []int32
			}
			bytes, err := Encode(&ErrTestStruct{}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})

		Convey("Should return error if struct string field doesn't contains length param", func() {
			type ErrTestStruct struct {
				Field string
			}
			bytes, err := Encode(&ErrTestStruct{Field: "Hello"}, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if struct tag length contains wrong value in nil value", func() {
			type ErrTestStruct struct {
				Field string `d2b:"length:1qwe"` //31
			}
			var d *ErrTestStruct
			bytes, err := Encode(d, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if we can't detect struct field array length", func() {
			type ErrTestStruct struct {
				Field [5]int
			}
			var d *ErrTestStruct
			bytes, err := Encode(d, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if we can't detect struct field length due to missing length tag", func() {
			type ErrTestStruct struct {
				Field *string
			}
			var d *ErrTestStruct
			bytes, err := Encode(d, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if unsupported type passed", func() {
			i := 0
			values := []interface{}{int(0), uint(0), []int32{}, [2]int{}, &i}
			for _, value := range values {
				bytes, err := Encode(value, binary.LittleEndian)
				So(err, ShouldNotBeNil)
				So(bytes, ShouldBeEmpty)
			}
		})
		Convey("Should return error if struct contains field of type pointer to unsupported type", func() {
			type ErrTestStruct struct {
				Field *int
			}
			var d = new(ErrTestStruct)
			bytes, err := Encode(d, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if array with bad elements encodes", func() {
			var data *[5]int
			bytes, err := Encode(data, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if struct slice with bad tag encodes", func() {
			type ErrTestStruct struct {
				Field int32 `d2b:"length:hello"`
			}
			var data *ErrTestStruct
			bytes, err := Encode(data, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if nil struct with slice field without length encodes", func() {
			type ErrTestStruct struct {
				Field []int32
			}
			var data *ErrTestStruct
			bytes, err := Encode(data, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
		Convey("Should return error if nil struct with slice field with bad element encodes", func() {
			type ErrTestStruct struct {
				Field []int `d2b:"length:5"`
			}
			var data *ErrTestStruct
			bytes, err := Encode(data, binary.LittleEndian)
			So(err, ShouldNotBeNil)
			So(bytes, ShouldBeEmpty)
		})
	})
}
