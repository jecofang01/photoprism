package query

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFilesByPath(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "Holiday")

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
	t.Run("files found - path starting with /", func(t *testing.T) {
		files, err := FilesByPath(10, 0, entity.RootOriginals, "/Holiday")

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(files))
	})
}

func TestExistingFiles(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := Files(1000, 0, "/", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(files))
	})
	t.Run("files found - includeMissing false", func(t *testing.T) {
		files, err := Files(1000, 0, "/", false)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}
		assert.LessOrEqual(t, 5, len(files))
	})
	t.Run("search for files path", func(t *testing.T) {
		files, err := Files(1000, 0, "Photos", true)

		t.Logf("files: %+v", files)

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, files)
	})
}

func TestFilesByUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		files, err := FilesByUID([]string{"ft8es39w45bnlqdw"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", files[0].FileName)
	})
	t.Run("no files found", func(t *testing.T) {
		files, err := FilesByUID([]string{"ft8es39w45bnlxxx"}, 100, 0)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(files))
	})
	t.Run("error", func(t *testing.T) {
		files, err := FilesByUID([]string{"ft8es39w45bnlxxx"}, -100, 0)

		assert.Error(t, err)
		assert.Equal(t, 0, len(files))
	})
}

func TestFileByPhotoUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByPhotoUID("pt9jtdre2lvl0y11")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "Germany/bridge.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByPhotoUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestVideoByPhotoUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := VideoByPhotoUID("pt9jtdre2lvl0yh0")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "1990/04/bridge2.mp4", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := VideoByPhotoUID("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByUID(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByUID("ft8es39w45bnlqdw")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByUID("111")

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestFileByHash(t *testing.T) {
	t.Run("files found", func(t *testing.T) {
		file, err := FileByHash("2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", file.FileName)
	})

	t.Run("no files found", func(t *testing.T) {
		file, err := FileByHash("111")

		assert.Error(t, err, "record not found")
		t.Log(file)
	})
}

func TestSetPhotoPrimary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.Equal(t, false, entity.FileFixturesExampleXMP.FilePrimary)

		err := SetPhotoPrimary("pt9jtdre2lvl0yh7", "ft2es49whhbnlqdn")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("no_file_uid", func(t *testing.T) {
		err := SetPhotoPrimary("pt9jtdre2lvl0yh7", "")

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("no_uid", func(t *testing.T) {
		err := SetPhotoPrimary("", "")

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("file missing", func(t *testing.T) {
		err := SetPhotoPrimary("pt9jtdre2lvl0y22", "")

		if err == nil {
			t.Fatal("error expected")
		}
		assert.Contains(t, err.Error(), "can't find primary file")
	})
}

func TestSetFileError(t *testing.T) {
	assert.Equal(t, "", entity.FileFixturesExampleXMP.FileError)

	SetFileError("ft2es49whhbnlqdn", "errorFromTest")

	//TODO How to assert
	//assert.Equal(t, true, entity.FileFixturesExampleXMP.FilePrimary)
}

func TestIndexedFiles(t *testing.T) {
	if err := entity.AddDuplicate(
		"Photo18.jpg",
		entity.RootSidecar,
		"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		661858,
		time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
	); err != nil {
		t.Fatal(err)
	}

	result, err := IndexedFiles()

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("INDEXED FILES: %#v", result)
}

func TestFileHashes(t *testing.T) {
	result, err := FileHashes()

	if err != nil {
		t.Fatal(err)
	}

	if len(result) < 3 {
		t.Fatalf("at least 3 file hashes expected")
	}

	t.Logf("FILE HASHES: %#v", result)
}

func TestRenameFile(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		err := RenameFile("xxx", "", "yyy", "yyy")

		if err == nil {
			t.Fatal(err)
		}
	})
	t.Run("success", func(t *testing.T) {
		assert.Equal(t, "2790/02/Photo01.xmp", entity.FileFixturesExampleXMP.FileName)
		assert.Equal(t, "/", entity.FileFixturesExampleXMP.FileRoot)
		err := RenameFile("/", "exampleXmpFile.xmp", "test-root", "yyy.jpg")

		if err != nil {
			t.Fatal(err)
		}
		//TODO how to assert?
		//assert.Equal(t, "", entity.FileFixturesExampleXMP.FileName)
	})

}
