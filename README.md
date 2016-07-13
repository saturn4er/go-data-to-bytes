#go-convert-to-bytes
This library will help you to convert your data to bytes array and your bytes array to data.

## Installation

    go get github.com/saturn4er/go-data-to-bytes

## Usage:

### Structure to bytes

    import "github.com/saturn4er/go-convert-to-bytes"
    type Test struct {
        A       string `bytes_length:"10"`
        B       string `bytes_length:"2"`
    }
    func main(){
        usefulData := Test{
            A: "hello",
            B: "world",
        }
        b, err := dtb.ConvertDataToBytes(usefulData, binary.LittleEndian)
        if err != nil {
            panic(err)   
        }
        fmt.Println(b) // [104 101 108 108 111 0 0 0 0 0 119 111] - Where first 10 bytes represent string "Hello", and last 2 bytes represent 2 characters of string "world"
    }
    
#### Available tags
    
 - bytes_length:"N"         - length of string in bytes
 - bytes_ignore:"true"      - ignore this field during conversion
 
### Array to bytes

    import "github.com/saturn4er/go-convert-to-bytes"
    type Test struct {
    	A string `bytes_length:"10"`
    	B string `bytes_length:"10"`
    }
    
    func main() {
    	usefulData := [2]Test{Test{A:"hello", B: "world"}, Test{A:"test", B:"secondtest"}}
    	b, err := dtb.ConvertDataToBytes(usefulData, binary.LittleEndian)
    	if err != nil {
    		panic(err)
    	}
    	fmt.Println(b) /* [104 101 108 108 111 0 0 0 0 0 - Hello
    			   119 111 114 108 100 0 0 0 0 0 - world
    			   116 101 115 116 0 0 0 0 0 0 - test
    			   115 101 99 111 110 100 116 101 115 116] - secondtest
    			*/
    }