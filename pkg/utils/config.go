package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func SetupConfig() (*Config, error) {
	conf := &Config{
		Init: false,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return conf, err
	}

	keyDir := filepath.Join(homeDir, ".gocrypt", "node")
	DHTPath := filepath.Join(homeDir, ".gocrypt", "DHT")
	pubPath := filepath.Join(keyDir, "PB")
	privPath := filepath.Join(keyDir, "PK")

	conf.homeDir = homeDir
	conf.DHTDir = DHTPath
	conf.KeyDir = keyDir

	_, err = os.Stat(keyDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(keyDir, 0755)
		if err != nil {
			return conf, err
		}
		log.Println("Directory created:", keyDir)
		os.WriteFile(DHTPath, nil, 0644)

		privKey, pubKey, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			return conf, err
		}

		marsPB, err := crypto.MarshalPublicKey(pubKey)
		if err != nil {
			return conf, err
		}
		os.WriteFile(pubPath, marsPB, 0644)

		marsPK, err := crypto.MarshalPrivateKey(privKey)
		if err != nil {
			return conf, err
		}
		err = os.WriteFile(privPath, marsPK, 0644)
		if err != nil {
			return conf, err
		}

		conf.Init = true
		fmt.Print("Successfully created config files!\n")

	} else {
		fmt.Print("Config already exists. Starting node from config!\n")
	}

	return conf, nil
}

func UpdateKeys(conf Config, pk crypto.PrivKey, pb crypto.PubKey) error {
	keyDir := conf.KeyDir
	pubPath := filepath.Join(keyDir, "pb.key")
	privPath := filepath.Join(keyDir, "pk.key")

	pbFile, err := os.OpenFile(pubPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer pbFile.Close()

	pkFile, err := os.OpenFile(privPath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer pkFile.Close()

	marPB, err := crypto.MarshalPublicKey(pb)
	if err != nil {
		return err
	}

	marPK, err := crypto.MarshalPrivateKey(pk)
	if err != nil {
		return err
	}

	_, err = pkFile.Write(marPK)
	if err != nil {
		return err
	}
	_, err = pbFile.Write(marPB)
	if err != nil {
		return err
	}

	fmt.Print("Successfully updated keys!\n")
	return nil
}

func ReadKeys(conf Config) (crypto.PubKey, crypto.PrivKey, error) {
	pubPath := filepath.Join(conf.KeyDir, "PB")
	privPath := filepath.Join(conf.KeyDir, "PK")

	marPB, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, err
	}

	marPK, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, err
	}

	pb, err := crypto.UnmarshalPublicKey(marPB)
	if err != nil {
		return nil, nil, err
	}

	pk, err := crypto.UnmarshalPrivateKey(marPK)
	if err != nil {
		return nil, nil, err
	}

	return pb, pk, nil
}
