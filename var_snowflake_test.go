package var_snowflake

import (
	"fmt"
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
		time.Sleep(time.Second)
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

