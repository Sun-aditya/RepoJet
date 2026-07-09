package workspace

import (
	"os"
	"testing"
)

func TestCreateAndRemoveWorkspace(t *testing.T) {
	path, err := Create()

	if err != nil {
		t.Fatalf("Create() returned an error: %v", err)
	}

	_, err = os.Stat(path)

	if err != nil {
		t.Fatalf("workspace directory was not created: %v", err)
	}

	err = Remove(path)

	if err != nil {
		t.Fatalf("Remove() returned an error: %v", err)
	}

	_, err = os.Stat(path)

	if !os.IsNotExist(err) {
		t.Errorf("workspace directory still exists after Remove()")
	}
}