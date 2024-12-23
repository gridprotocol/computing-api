package keystore

type KeyType = byte

const (
	// RSA is an enum for the supported RSA key type
	RSA KeyType = iota
	// Secp256k1 is an enum for the supported Secp256k1 key type
	Secp256k1
	BLS
	PDP
	Ed25519
	Hmac
)

// KeyStore is used for storing secret keys
type KeyStore interface {
	// List lists all the keys stored in the KeyStore
	List() ([]string, error)
	// Get gets a key out of keystore use its name and password
	Get(string, string) (KeyInfo, error)
	// Put saves a key info under given name and password
	Put(string, string, KeyInfo) error
	// Delete removes a key from keystores
	Delete(string, string) error
	// check if a name exist in the keystore
	Exist(name string) (bool, error)

	Close() error
}
