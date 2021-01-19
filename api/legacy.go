package api

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bip39 "github.com/cosmos/go-bip39"
)

var (
	addressSuffix          = "address"
	infoSuffix             = "info"
	defaultBIP39Passphrase = ""
	FullFundraiserPath     = "44'/118'/0'/0/0"
)

// LegacyKeybase is implemented by the legacy keybase implementation.
type LegacyKeybase interface {
	List() ([]keyring.Info, error)
	Get(name string) (Info, error)
	CreateKey(name, mnemonic, passwd string) (info Info, err error)
	Update(name, oldpass string, getNewpass func() (string, error)) error
	Delete(name, passphrase string, skipPass bool) error
	Export(name string) (armor string, err error)
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)
	ExportPubKey(name string) (armor string, err error)
	Sign(name, passphrase string, msg []byte) (sig []byte, pub cryptotypes.PubKey, err error)
	Close() error
}

// NewLegacy creates a new instance of a legacy keybase.
func NewLegacy(name, dir string, opts ...keyring.KeybaseOption) (LegacyKeybase, error) {
	if err := tmos.EnsureDir(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create Keybase directory: %s", err)
	}

	db, err := sdk.NewLevelDB(name, dir)
	if err != nil {
		return nil, err
	}

	return newDBKeybase(db), nil
}

var _ LegacyKeybase = dbKeybase{}

// dbKeybase combines encryption and storage implementation to provide a
// full-featured key manager.
//
// NOTE: dbKeybase will be deprecated in favor of keyringKeybase.
type dbKeybase struct {
	db dbm.DB
}

// newDBKeybase creates a new dbKeybase instance using the provided DB for
// reading and writing keys.
func newDBKeybase(db dbm.DB) dbKeybase {
	return dbKeybase{
		db: db,
	}
}

// List returns the keys from storage in alphabetical order.
func (kb dbKeybase) List() ([]keyring.Info, error) {
	var res []keyring.Info

	iter, err := kb.db.Iterator(nil, nil)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())

		// need to include only keys in storage that have an info suffix
		if strings.HasSuffix(key, infoSuffix) {
			info, err := unmarshalInfo(iter.Value())
			if err != nil {
				return nil, err
			}

			res = append(res, info)
		}
	}

	return res, nil
}

// Get returns the public information about one key.
func (kb dbKeybase) Get(name string) (Info, error) {
	bs, err := kb.db.Get(infoKey(name))
	if err != nil {
		return nil, err
	}

	if len(bs) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, name)
	}

	return unmarshalInfo(bs)
}

// TEMPORARY METHOD UNTIL WE FIGURE OUT USER FACING HD DERIVATION API
func (kb dbKeybase) CreateKey(name, mnemonic, passwd string) (info Info, err error) {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		err = fmt.Errorf("recovering only works with 12 word (fundraiser) or 24 word mnemonics, got: %v words", len(words))
		return
	}
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return
	}
	info, err = kb.persistDerivedKey(seed, passwd, name, FullFundraiserPath)
	return
}

// Delete removes key forever, but we must present the
// proper passphrase before deleting it (for security).
// A passphrase of 'yes' is used to delete stored
// references to offline and Ledger / HW wallet keys
func (kb dbKeybase) Delete(name, passphrase string, skipPass bool) error {
	// verify we have the proper password before deleting
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	if linfo, ok := info.(localInfo); ok && !skipPass {
		if _, _, err = crypto.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase); err != nil {
			return err
		}
	}
	kb.db.DeleteSync(addrKey(info.GetAddress()))
	kb.db.DeleteSync(infoKey(name))
	return nil
}

// Update changes the passphrase with which an already stored key is
// encrypted.
//
// oldpass must be the current passphrase used for encryption,
// getNewpass is a function to get the passphrase to permanently replace
// the current passphrase
func (kb dbKeybase) Update(name, oldpass string, getNewpass func() (string, error)) error {
	info, err := kb.Get(name)
	if err != nil {
		return err
	}
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		key, _, err := crypto.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, oldpass)
		if err != nil {
			return err
		}
		newpass, err := getNewpass()
		if err != nil {
			return err
		}
		kb.writeLocalKey(key, name, newpass)
		return nil
	default:
		return fmt.Errorf("locally stored key required")
	}
}

func (kb *dbKeybase) persistDerivedKey(seed []byte, passwd, name, fullHdPath string) (info Info, err error) {
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, fullHdPath)
	if err != nil {
		return
	}

	// if we have a password, use it to encrypt the private key and store it
	// else store the public key only

	info = kb.writeLocalKey(&secp256k1.PrivKey{derivedPriv}, name, passwd)

	return
}

// ExportPrivateKeyObject returns a PrivKey object given the key name and
// passphrase. An error is returned if the key does not exist or if the Info for
// the key is invalid.
func (kb dbKeybase) ExportPrivateKeyObject(name string, passphrase string) (types.PrivKey, error) {
	info, err := kb.Get(name)
	if err != nil {
		return nil, err
	}

	var priv types.PrivKey

	switch i := info.(type) {
	case localInfo:
		linfo := i
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return nil, err
		}

		priv, _, err = crypto.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, err
		}

	case ledgerInfo, offlineInfo, multiInfo:
		return nil, errors.New("only works on local private keys")
	}

	return priv, nil
}

func (kb dbKeybase) Export(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	return crypto.ArmorInfoBytes(bz), nil
}

// ExportPubKey returns public keys in ASCII armored format. It retrieves a Info
// object by its name and return the public key in a portable format.
func (kb dbKeybase) ExportPubKey(name string) (armor string, err error) {
	bz, err := kb.db.Get(infoKey(name))
	if err != nil {
		return "", err
	}

	if bz == nil {
		return "", fmt.Errorf("no key to export with name %s", name)
	}

	info, err := unmarshalInfo(bz)
	if err != nil {
		return
	}

	return crypto.ArmorPubKeyBytes(info.GetPubKey().Bytes(), string(info.GetAlgo())), nil
}

// ExportPrivKey returns a private key in ASCII armored format.
// It returns an error if the key does not exist or a wrong encryption passphrase
// is supplied.
func (kb dbKeybase) ExportPrivKey(name string, decryptPassphrase string,
	encryptPassphrase string) (armor string, err error) {
	priv, err := kb.ExportPrivateKeyObject(name, decryptPassphrase)
	if err != nil {
		return "", err
	}

	info, err := kb.Get(name)
	if err != nil {
		return "", err
	}

	return crypto.EncryptArmorPrivKey(priv, encryptPassphrase, string(info.GetAlgo())), nil
}

func (kb dbKeybase) Sign(name, passphrase string, msg []byte) (sig []byte, pub cryptotypes.PubKey, err error) {
	info, err := kb.Get(name)
	if err != nil {
		return
	}
	var priv cryptotypes.PrivKey
	switch info.(type) {
	case localInfo:
		linfo := info.(localInfo)
		if linfo.PrivKeyArmor == "" {
			err = fmt.Errorf("private key not available")
			return
		}
		priv, _, err = crypto.UnarmorDecryptPrivKey(linfo.PrivKeyArmor, passphrase)
		if err != nil {
			return nil, nil, err
		}
	}
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub = priv.PubKey()
	return sig, pub, nil
}

// Close the underlying storage.
func (kb dbKeybase) Close() error { return kb.db.Close() }

func infoKey(name string) []byte { return []byte(fmt.Sprintf("%s.%s", name, infoSuffix)) }

func addrKey(address sdk.AccAddress) []byte {
	return []byte(fmt.Sprintf("%s.%s", address.String(), addressSuffix))
}

func (kb dbKeybase) writeLocalKey(priv cryptotypes.PrivKey, name, passphrase string) Info {

	// encrypt private key using passphrase
	privArmor := crypto.EncryptArmorPrivKey(priv, passphrase, priv.Type())
	// make Info
	pub := priv.PubKey()
	info := newLocalInfo(name, pub, privArmor, hd.PubKeyType(priv.Type()))
	kb.writeInfo(info, name)
	return info
}

func (kb dbKeybase) writeInfo(info Info, name string) {
	// write the info by key
	key := infoKey(name)
	kb.db.SetSync(key, writeInfo(info))
	// store a pointer to the infokey by address for fast lookup
	kb.db.SetSync(addrKey(info.GetAddress()), key)
}

// encoding info
func writeInfo(i Info) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(i)
}
