// +build e2e

package ftpclient

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	notNilAssertMessage = "expecting not-nil"
	nilAssertMessage    = "expecting nil"

	envLogin    = "LOGIN"
	envPassword = "PASSWORD"
	envFilesDir = "FILES_DIR"

	newDirName         = "new_dir"
	notExistingDirName = "not_dir_file"

	newFileName         = "new_file"
	renameToName        = "after_rename"
	notExistingFileName = "not_existing_file"
)

var c *Client

func TestNewClient(t *testing.T) {
	var err error
	c, err = NewClient()

	assert.NotNil(t, c, notNilAssertMessage)
	assert.Nil(t, err, nilAssertMessage)
}

func TestLogin(t *testing.T) {
	err := c.Login(os.Getenv(envLogin), os.Getenv(envPassword))
	assert.Nil(t, err, nilAssertMessage)
}

func TestList(t *testing.T) {
	list, err := c.List()
	assert.Equal(t, true, strings.Contains(list, os.Getenv(envFilesDir)))
	assert.Nil(t, err, nilAssertMessage)
}

func TestChangeDir(t *testing.T) {
	err := c.ChangeDir(os.Getenv(envFilesDir))
	assert.Nil(t, err, nilAssertMessage)
}

func TestPwd(t *testing.T) {
	expectedDir := "/files"
	dir, err := c.PWD()
	assert.Equal(t, expectedDir, dir, "")
	assert.Nil(t, err, nilAssertMessage)
}

func TestMakeDir(t *testing.T) {
	err := c.MakeDir(newDirName)
	assert.Nil(t, err, nilAssertMessage)
}

func TestRemoveDir(t *testing.T) {
	err := c.RemoveDir(notExistingDirName)
	assert.NotNil(t, err, notNilAssertMessage)

	err = c.RemoveDir(newDirName)
	assert.Nil(t, err, nilAssertMessage)
}

func TestRenameFrom(t *testing.T) {
	err := c.RenameFrom(newFileName)
	assert.Nil(t, err, nilAssertMessage)
}

func TestRenameTo(t *testing.T) {
	err := c.RenameTo(renameToName)
	assert.Nil(t, err, nilAssertMessage)
}

func TestDelete(t *testing.T) {
	err := c.Delete(notExistingFileName)
	assert.NotNil(t, err, notNilAssertMessage)

	err = c.Delete(renameToName)
	assert.Nil(t, err, nilAssertMessage)
}

func TestSetBinaryType(t *testing.T) {
	err := c.SetBinaryType()
	assert.Nil(t, err, nilAssertMessage)
}

func TestQuit(t *testing.T) {
	err := c.Quit()
	assert.Nil(t, err, nilAssertMessage)
}
