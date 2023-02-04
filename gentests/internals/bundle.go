package internals

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/CopilotIQ/opa-tests/gentests"
	"io"
	"os"
	"path/filepath"
)

var (
	log = gentests.Log
)

// CreateBundle takes a JSON Manifest and the path to a directory containing OPA Policies (
// Rego files) and then generates an archive (`.tar.gz`) file according to OPA Bundle rules.
// It returns the full path to the temporary file.
func CreateBundle(manifestPath string, srcDir string) (string, error) {
	var manifest = gentests.ReadManifest(manifestPath)
	if manifest == nil {
		return "", fmt.Errorf("cannot load manifest %s", manifestPath)
	}
	pattern := fmt.Sprintf("bundle-%s-*.tar.gz", manifest.Revision)
	gzFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer gzFile.Close()
	log.Debug("generating bundle to %s", gzFile.Name())

	// gzip writer
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	// tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// TODO: walk the subtree (instead of just the directory) and modify the test names to
	// 		 reflect the position in the subtree using WalkDir(root string, fn fs.WalkDirFunc)
	files, err := filepath.Glob(filepath.Join(srcDir, gentests.PoliciesGlob))
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("empty policies directory %s", srcDir)
	}
	log.Debug("found Rego files: %s", files)

	// Adding the .manifest from the Manifest file
	tmpManifest, err := createTempManifest(manifestPath)
	if err != nil {
		return "", fmt.Errorf("could not create temporary .manifest: %v", err)
	}
	files = append(files, tmpManifest)
	err = tarFiles(files, tarWriter)
	if err != nil {
		return "", fmt.Errorf("could not create tarball from Rego files: %v", err)
	}
	return gzFile.Name(), nil
}

func createTempManifest(manifestPath string) (string, error) {
	srcFile, _ := os.Open(manifestPath)
	defer srcFile.Close()
	destFile, _ := os.Create(os.TempDir() + "/.manifest")
	defer destFile.Close()

	_, err := io.Copy(destFile, srcFile)
	if err != nil {
		log.Error("could not make a temporary copy of %s: %v", manifestPath, err)
		return "", err
	}
	return destFile.Name(), nil
}

func tarFiles(files []string, tarWriter *tar.Writer) error {
	// To avoid resource leaks, defer is called outside the for loop.
	var filesToClose = make([]*os.File, 0)
	defer func() {
		for _, f := range filesToClose {
			err := f.Close()
			if err != nil {
				log.Error("cannot close %s: %v", f.Name(), err)
			}
		}
	}()

	for _, file := range files {
		// open file
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		if err != nil {
			return err
		}

		// write file to tar
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Base(file)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}
	}
	return nil
}
