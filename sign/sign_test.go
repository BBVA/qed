package sign

import (
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func testSign(t *testing.T, signer Signer) {

	message := []byte("send reinforcements, we're going to advance")

	sig, _ := signer.Sign(message)
	result, _ := signer.Verify(message, sig)

	assert.True(t, result, "Must be verified")

}
func TestEdSign(t *testing.T) { testSign(t, NewEd25519Signer()) }

func syncBenchmark(b *testing.B, signer Signer, iterations int) {

	b.N = iterations
	for i := 0; i < b.N; i++ {
		signer.Sign([]byte(fmt.Sprintf("send reinforcements, we're going to advance %d", b.N)))
	}

}

func asyncBenchmark(b *testing.B, signer Signer, numRoutines, iterations int) {

	data := make(chan []byte)
	close := make(chan bool)

	for i := 1; i <= numRoutines; i++ {
		go func(i int, data chan []byte, close chan bool) {
			for {
				select {
				case msg := <-data:
					signer.Sign([]byte(fmt.Sprintf("send reinforcements, we're going to advance %s", msg)))
				case <-close:
					return
				}
			}
		}(i, data, close)
	}

	b.N = iterations

	for i := 1; i <= b.N; i++ {
		data <- []byte(fmt.Sprintf("data(%d)", i))
	}

	close <- true

}

func BenchmarkEd(b *testing.B) {

	for _, iterations := range []int{1e3, 1e4, 1e5} {
		for _, numRoutines := range []int{1e1, 1e2, 1e3} {
			b.Run(fmt.Sprintf("routines-%d", numRoutines), func(b *testing.B) {
				asyncBenchmark(b, NewEd25519Signer(), numRoutines, iterations)
			})
		}

		b.Run(fmt.Sprintf(""), func(b *testing.B) {
			syncBenchmark(b, NewEd25519Signer(), iterations)
		})
	}

}
