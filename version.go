package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	bolt "go.etcd.io/bbolt"
)

func addOrUpdateVersion(tx *bolt.Tx, bytes []byte, version, downloadLink, infoKey string) error {
	composerJson := map[string]interface{}{}

	if !matchVersionConstraint(version) {
		return errors.New(fmt.Sprintf("version '%s' does not match regex ^v?[0-9]+\\.[0-9]+\\.[0-9]+(\\-[A-z]+)?$", version))
	}

	if err := json.Unmarshal(bytes, &composerJson); err != nil {
		return err
	}

	return addOrUpdateVersionDirect(tx, composerJson, downloadLink, version, infoKey)
}

func deleteVersion(tx *bolt.Tx, saveTag string) error {
	saveTag = "info--" + saveTag

	bucket := tx.Bucket([]byte("packages"))

	versionKey := bucket.Get([]byte(saveTag))

	if versionKey == nil {
		return nil
	}

	err := bucket.Delete([]byte(saveTag))
	if err != nil {
		return err
	}

	err = bucket.Delete(versionKey)
	if err != nil {
		return err
	}

	return nil
}

func matchVersionConstraint(version string) bool {
	var digitCheck = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(\-[A-z]+)?$`)
	return digitCheck.MatchString(version)
}

func addOrUpdateVersionDirect(tx *bolt.Tx, composerJson map[string]interface{}, downloadLink, version, infoKey string) error {
	packageName := composerJson["name"].(string)

	composerJson["dist"] = map[string]string{
		"url":  downloadLink,
		"type": "zip",
	}

	composerJson["version"] = version

	key := fmt.Sprintf("packages--%s|%s", packageName, version)

	bucket := tx.Bucket([]byte("packages"))

	if err := bucket.Put([]byte("info--"+infoKey), []byte(key)); err != nil {
		return err
	}

	composerJsonData, _ := json.Marshal(composerJson)

	return bucket.Put([]byte(key), composerJsonData)
}
