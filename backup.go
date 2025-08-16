package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func backupWorker() {
	ticker := time.NewTicker(time.Duration(config.BackupInterval) * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		Info("Backup process started...")
		path, err := CreateZIPBackup()
		if err != nil {
			Error("can't create backup file", "err", err.Error())

			continue
		}

		if err := S3Upload(path); err != nil {
			Error("can't upload backup file", "err", err.Error())
		}
		Info("Backup was successful.")
	}
}

func S3Upload(backupPath string) error {
	client, err := minio.New(config.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.S3AccessKeyID, config.S3SecretKey, ""),
		Region: config.S3Region,
		Secure: config.S3Secure,
	})
	if err != nil {
		return err
	}

	file, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = client.PutObject(
		context.Background(),
		config.S3BucketName,
		backupPath,
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		return err
	}

	if err := os.Remove(backupPath); err != nil {
		return err
	}

	return nil
}

func CreateZIPBackup() (string, error) {
	backupPath := path.Join(config.WorkingDirectory,
		fmt.Sprintf("alienos-backup-%s.zip", time.Now().Format("2006-January-02")))
	file, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	if err := filepath.Walk(config.WorkingDirectory, walker); err != nil {
		return "", err
	}

	return backupPath, nil
}
