package main

import (
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/VoycerAG/lego/acme"
)

// Account represents a users local saved credentials
type Account struct {
	Email        string `json:"email"`
	key          *rsa.PrivateKey
	Registration *acme.RegistrationResource `json:"registration"`

	conf *Configuration
}

// NewAccount creates a new account for an email address
func NewAccount(email string, conf *Configuration) *Account {
	accKeysPath := conf.AccountKeysPath(email)
	// TODO: move to function in configuration?
	accKeyPath := accKeysPath + string(os.PathSeparator) + email + ".key"
	if err := checkFolder(accKeysPath); err != nil {
		logger().Fatalf("Could not check/create directory for account %s: %v", email, err)
	}

	var privKey *rsa.PrivateKey
	if _, err := os.Stat(accKeyPath); os.IsNotExist(err) {
		logger().Printf("No key found for account %s. Generating a %v bit key.", email, conf.RsaBits())
		privKey, err = generateRsaKey(conf.RsaBits(), accKeyPath)
		if err != nil {
			logger().Fatalf("Could not generate RSA private account key for account %s: %v", email, err)
		}
		logger().Printf("Saved key to %s", accKeyPath)
	} else {
		privKey, err = loadRsaKey(accKeyPath)
		if err != nil {
			logger().Fatalf("Could not load RSA private key from file %s: %v", accKeyPath, err)
		}
	}

	accountFile := path.Join(conf.AccountPath(email), "account.json")
	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privKey, conf: conf}
	}

	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		logger().Fatalf("Could not load file for account %s -> %v", email, err)
	}

	var acc Account
	err = json.Unmarshal(fileBytes, &acc)
	if err != nil {
		logger().Fatalf("Could not parse file for account %s -> %v", email, err)
	}

	acc.key = privKey
	acc.conf = conf

	return &acc
}

/** Implementation of the acme.User interface **/

// GetEmail returns the email address for the account
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() *rsa.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration
func (a *Account) GetRegistration() *acme.RegistrationResource {
	return a.Registration
}

/** End **/

// Save the account to disk
func (a *Account) Save() error {
	jsonBytes, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(a.conf.AccountPath(a.Email), "account.json"), jsonBytes, 0700)

}
