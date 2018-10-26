package publish

import "testing"

func testSign(t *testing.T) {
	// signer := sign.NewEd25519Signer()

	// Check the body response
	// signedSnapshot := &SignedSnapshot{}

	// json.Unmarshal([]byte(rr.Body.String()), signedSnapshot)

	// if !bytes.Equal(signedSnapshot.Snapshot.HyperDigest, []byte{0x1}) {
	// 	t.Errorf("HyperDigest is not consistent: %s", signedSnapshot.Snapshot.HyperDigest)
	// }

	// if !bytes.Equal(signedSnapshot.Snapshot.HistoryDigest, []byte{0x0}) {
	// 	t.Errorf("HistoryDigest is not consistent %s", signedSnapshot.Snapshot.HistoryDigest)
	// }

	// if signedSnapshot.Snapshot.Version != 0 {
	// 	t.Errorf("Version is not consistent")
	// }

	// if !bytes.Equal(signedSnapshot.Snapshot.Event, []byte("this is a sample event")) {
	// 	t.Errorf("Event is not consistent ")
	// }

	// signature, err := signer.Sign([]byte(fmt.Sprintf("%v", signedSnapshot.Snapshot)))

	// if !bytes.Equal(signedSnapshot.Signature, signature) {
	// 	t.Errorf("Signature is not consistent")
	// }
}

func testPublisher(t *testing.T) {

}
