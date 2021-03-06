package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mholt/archiver"
)

func cleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func copyPath(src, dst string, dryRun bool) error {
	return filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filePath = filepath.ToSlash(filePath) // fix Windows slashes
		filename := strings.Trim(strings.TrimPrefix(filePath, src), "/")
		dstFilePath := filepath.Join(dst, filename)
		if dryRun {
			if info.IsDir() {
				log.Printf("make path %s", dstFilePath)
				return nil
			}
			if !info.Mode().IsRegular() {
				log.Printf("'%s' is not a regular file, skipping", filePath)
				return nil
			}
			log.Printf("copy %s -> %s", filePath, dstFilePath)
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(dstFilePath, os.ModePerm)
		}
		if !info.Mode().IsRegular() {
			log.Printf("'%s' is not a regular file, skipping", filePath)
			return nil
		}
		return copyFile(filePath, dstFilePath)
	})
}

func copyFile(srcFile string, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func GetBackupsToDelete(backups []string, keep int) []string {
	type parsedBackup struct {
		name string
		time time.Time
	}
	backupList := []parsedBackup{}
	for _, backupName := range backups {
		t, err := time.Parse(BackupTimeFormat, strings.TrimSuffix(backupName, ".tar"))
		if err == nil {
			backupList = append(backupList, parsedBackup{
				name: backupName,
				time: t,
			})
		}
	}
	sort.SliceStable(backupList, func(i, j int) bool {
		return backupList[i].time.Before(backupList[j].time)
	})
	result := []string{}
	if len(backupList) > keep {
		for i := 0; i < len(backupList)-keep; i++ {
			result = append(result, backupList[i].name)
		}
	}
	return result
}

func getArchiveWriter(format string, level int) (archiver.Writer, error) {
	switch format {
	case "tar":
		return &archiver.Tar{}, nil
	case "lz4":
		return &archiver.TarLz4{CompressionLevel: level, Tar: archiver.NewTar()}, nil
	case "bzip2":
		return &archiver.TarBz2{CompressionLevel: level, Tar: archiver.NewTar()}, nil
	case "gzip":
		return &archiver.TarGz{CompressionLevel: level, Tar: archiver.NewTar()}, nil
	case "sz":
		return &archiver.TarSz{Tar: archiver.NewTar()}, nil
	case "xz":
		return &archiver.TarXz{Tar: archiver.NewTar()}, nil
	}
	return nil, fmt.Errorf("wrong compression_format, supported: 'lz4', 'bzip2', 'gzip', 'sz', 'xz'")
}

func getExtension(format string) string {
	switch format {
	case "tar":
		return "tar"
	case "lz4":
		return "tar.lz4"
	case "bzip2":
		return "tar.bz2"
	case "gzip":
		return "tar.gz"
	case "sz":
		return "tar.sz"
	case "xz":
		return "tar.xz"
	}
	return ""
}

func getArchiveReader(format string) (archiver.Reader, error) {
	switch format {
	case "tar":
		return archiver.NewTar(), nil
	case "lz4":
		return archiver.NewTarLz4(), nil
	case "bzip2":
		return archiver.NewTarBz2(), nil
	case "gzip":
		return archiver.NewTarGz(), nil
	case "sz":
		return archiver.NewTarSz(), nil
	case "xz":
		return archiver.NewTarXz(), nil
	}
	return nil, fmt.Errorf("wrong compression_format, supported: 'tar', 'lz4', 'bzip2', 'gzip', 'sz', 'xz'")
}
