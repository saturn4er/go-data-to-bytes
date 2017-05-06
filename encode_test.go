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
			FieldA string     `d2b:"length:2"` //31
			FieldB string     `d2b:"length:2"` // 33
			Slice  []int32    `d2b:"length:2"`
			A      CustomType `d2b:"length:2"`
			QQ     *int32
			B      int16  // 2
			C      int32  // 4 6
			D      int64  // 8 14
			E      uint8  // 1 15
			F      uint16 // 2 17
			G      uint32 // 4 21
			H      uint64 // 8 29
			EE     string `d2b:"-"`
			I      int8
			Q      *uint8
			Z      [0]int32 `d2b:"length:1"`
		}
		Convey("Should return empty byte array if nil passed", func() {
			var data *TestStruct1
			bytes, err := Encode(data, binary.LittleEndian)
			if err != nil {
				t.Error(err)
				return
			}
			So(err, ShouldBeNil)
			So(bytes, ShouldHaveLength, 59)
		})
		Convey("Should return valid byte array", func() {
			data := &TestStruct1{
				FieldA: "Hello",
				FieldB: "World",
				QQ: new(int32),
			}
			bytes, err := Encode(data, binary.LittleEndian)
			if err != nil {
				t.Error(err)
				return
			}
			So(err, ShouldBeNil)
			So(bytes, ShouldHaveLength, 59)
			So(string(bytes[:2]), ShouldEqual, "He")
			So(string(bytes[2:4]), ShouldEqual, "Wo")
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
		Convey("Should return error if struct contains array field with type, that's not valid for marshalling 2", func() {
			type ErrTestStruct struct {
				Field [0]int `d2b:"length:4"` //31
			}
			bytes, err := Encode(&ErrTestStruct{}, binary.LittleEndian)
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
	})
}
