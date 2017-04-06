package filefixture

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"

	humanize "github.com/dustin/go-humanize"
	"github.com/google/jsonapi"
)

// List is a list of FileFixtures, used for index action
type List struct {
	Fixtures []*FileFixture
}

// FileFixture represents a single file fixture
type FileFixture struct {
	ID             string   `jsonapi:"primary,file_fixture_structureds"`
	Name           string   `jsonapi:"attr,name"`
	CurrentVersion *Version `jsonapi:"relation,current_version"`
	Versions       *Version `jsonapi:"relation,versions"`
	CreatedAt      string   `jsonapi:"attr,created_at"`
	UpdatedAt      string   `jsonapi:"attr,updated_at"`
}

// Version respresents a concrete version of a file fixture
type Version struct {
	ID               string `jsonapi:"primary,file_fixture_version_structureds"`
	OriginalHash     string `jsonapi:"attr,hash"`
	OriginalFileSize string `jsonapi:"attr,file_size"`
	FieldNames       string `jsonapi:"attr,field_names"`
	CreatedAt        string `jsonapi:"attr,created_at"`
	UpdatedAt        string `jsonapi:"attr,updated_at"`
}

// UnmarshalFileFixtures unmarshals a list of FileFixture records
func UnmarshalFileFixtures(input io.Reader) (List, error) {
	items, err := jsonapi.UnmarshalManyPayload(input, reflect.TypeOf(new(FileFixture)))
	if err != nil {
		return List{}, err
	}

	result := List{}

	for _, item := range items {
		fixture, ok := item.(*FileFixture)
		if !ok {
			return List{}, fmt.Errorf("Type assertion failed")
		}

		result.Fixtures = append(result.Fixtures, fixture)
	}

	return result, nil
}

// UnmarshalFileFixture unmarshals a single FileFixture record
func UnmarshalFileFixture(input io.Reader) (*FileFixture, error) {
	fixture := new(FileFixture)
	err := jsonapi.UnmarshalPayload(input, fixture)
	if err != nil {
		return new(FileFixture), err
	}

	return fixture, nil
}

// FindByName look up a FileFixture by name in List
func (fileFixtures List) FindByName(name string) FileFixture {
	for _, fileFixture := range fileFixtures.Fixtures {
		if fileFixture.Name == name {
			return *fileFixture
		}
	}

	return FileFixture{}
}

// ShowName displays the name, id and hash of a filefixture
func ShowName(input io.Reader) {
	items, err := jsonapi.UnmarshalManyPayload(input, reflect.TypeOf(new(FileFixture)))

	if err != nil {
		log.Fatal(err)
	}

	for _, item := range items {
		fixture, _ := item.(*FileFixture)

		fmt.Printf("* %s (ID: %s, Content-Hash: %s)\n", fixture.Name, fixture.ID, fixture.CurrentVersion.OriginalHash)
	}
}

// ShowDetails print out details of a file fixture, including its current version
func ShowDetails(fileFixture *FileFixture) {
	// TODO where to move this? shouldn't this be already done earlier?
	bytes, err := strconv.ParseUint(fileFixture.CurrentVersion.OriginalFileSize, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name:            %s\n", fileFixture.Name)
	fmt.Printf("UID:             %s\n", fileFixture.ID)
	fmt.Printf("Current Version: %s\n", fileFixture.CurrentVersion.ID)
	fmt.Printf("  SHA256 Hash:   %s\n", fileFixture.CurrentVersion.OriginalHash)
	fmt.Printf("  Size:          %s\n", humanize.Bytes(bytes))
	fmt.Printf("  Created:       %s\n", fileFixture.CurrentVersion.CreatedAt)
}
