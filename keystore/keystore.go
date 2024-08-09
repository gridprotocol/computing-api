package keystore

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
	//"github.com/grid/test-ks/types"
)

var _ KeyStore = (*keyStore)(nil)

type keyStore struct {
	path string
}

// create a repo to store keyfile
// todo: create a file for verifing password before reopen
func NewKeyStore(path string) (KeyStore, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	// create a file for test password

	return &keyStore{
		path,
	}, nil
}

// StorePrivateKey encrypt the privatekey by password and then store it in keystore
func (k keyStore) Put(name, auth string, info KeyInfo) error {
	sData, err := json.Marshal(info)
	if err != nil {
		return err
	}

	key := &Key{
		Address:     name,
		SecretValue: sData,
	}

	keyjson, err := encryptKey(key, auth, StandardScryptN, StandardScryptP)
	if err != nil {
		return err
	}

	path := joinPath(k.path, name)
	_, err = os.Stat(path)
	if err == nil {
		// exist
		return nil
	}

	return writeKeyFile(path, keyjson)
}

func (k *keyStore) Get(name, auth string) (KeyInfo, error) {
	var res KeyInfo

	path := joinPath(k.path, name)

	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return res, err
	}

	key, err := decryptKey(keyjson, auth)
	if err != nil {
		return res, err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if strings.Compare(key.Address, name) != 0 {
		return res, xerrors.Errorf("key content mismatch: have peer %x, want %x", key.Address, name)
	}

	err = json.Unmarshal(key.SecretValue, &res)
	if err != nil {
		return res, xerrors.Errorf("decoding key '%s': %w", name, err)
	}

	return res, nil
}

func (k *keyStore) List() ([]string, error) {
	dir, err := os.Open(k.path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(files))
	for _, f := range files {
		keys = append(keys, string(f.Name()))
	}
	return keys, nil
}

// check if an name exists
func (k *keyStore) Exist(name string) (bool, error) {
	l, err := k.List()
	if err != nil {
		return false, err
	}

	for _, v := range l {
		if v == name {
			return true, nil
		}
	}

	return false, nil
}

func (k *keyStore) Delete(name, auth string) error {
	_, err := k.Get(name, auth)
	if err != nil {
		return err
	}

	keyPath := joinPath(k.path, name)

	err = os.Remove(keyPath)
	if err != nil {
		return err
	}

	return nil
}

func (k *keyStore) Close() error {
	return nil
}

func joinPath(dir string, filename string) (path string) {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(dir, filename)
}

func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	err := os.MkdirAll(filepath.Dir(file), dirPerm)
	if err != nil {
		return "", err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

func writeKeyFile(file string, content []byte) error {
	name, err := writeTemporaryKeyFile(file, content)
	if err != nil {
		return err
	}
	return os.Rename(name, file)
}

func LoadKeyFile(password, path string) (string, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	key, err := decryptKey(keyjson, password)
	if err != nil {
		return "", err
	}

	sk := key.SecretValue
	var res KeyInfo
	err = json.Unmarshal(key.SecretValue, &res)
	if err == nil {
		sk = res.SecretKey
	}

	enc := make([]byte, len(sk)*2)
	hex.Encode(enc, sk)
	return string(enc), nil
}
