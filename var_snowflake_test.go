package var_snowflake

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestNode_Generate(t *testing.T) {
	generator, err := NewNode(Epoch20200101, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	for j := 0; j < 3; j++ {
		for i := 0; i < 5; i++ {
			id := generator.Generate()
			fmt.Printf("%s = %s = %d\n", id.Base64(), id.Base2(), id.Int64())
		}
		fmt.Println("-----------------")
		time.Sleep(1 * time.Second)
	}
}

func BenchmarkNode_Generate(b *testing.B) {
	generator, _ := NewNode(Epoch20200101, 1)
	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = generator.Generate()
	}
}

func TestBits1(t *testing.T) {

	mask := int64(-1 ^ (-1 << 29))

	now := time.Since(Epoch20200101).Nanoseconds() / 1e9
	fmt.Println(strconv.FormatInt(now, 2))
	fmt.Println(strconv.FormatInt(now & mask, 2))
	fmt.Println(strconv.FormatInt(^now & mask, 2))
}
