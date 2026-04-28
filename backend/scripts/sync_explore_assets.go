package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"image-play/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	if cfg.OSSBucket == "" || cfg.OSSEndpoint == "" || cfg.OSSAccessKeyID == "" || cfg.OSSAccessKeySecret == "" {
		log.Fatal("OSS config missing. Check config.yaml or env vars.")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	client, err := oss.New(cfg.OSSEndpoint, cfg.OSSAccessKeyID, cfg.OSSAccessKeySecret)
	if err != nil {
		log.Fatalf("oss new: %v", err)
	}

	bucket, err := client.Bucket(cfg.OSSBucket)
	if err != nil {
		log.Fatalf("oss bucket: %v", err)
	}

	marker := ""
	prefix := "explore/"
	inserted := 0
	skipped := 0

	for {
		res, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix), oss.MaxKeys(100))
		if err != nil {
			log.Fatalf("list objects: %v", err)
		}

		for _, obj := range res.Objects {
			if !strings.HasSuffix(strings.ToLower(obj.Key), ".jpg") && !strings.HasSuffix(strings.ToLower(obj.Key), ".jpeg") {
				continue
			}

			imageURL := fmt.Sprintf("https://%s.%s/%s", cfg.OSSBucket, cfg.OSSEndpoint, obj.Key)

			var exists int
			err := db.QueryRow("SELECT 1 FROM explore_assets WHERE image_url = $1", imageURL).Scan(&exists)
			if err == nil {
				skipped++
				continue
			}

			_, err = db.Exec(
				"INSERT INTO explore_assets (image_url, scene_key, prompt) VALUES ($1, $2, $3)",
				imageURL, "", "",
			)
			if err != nil {
				log.Printf("insert failed for %s: %v", imageURL, err)
				continue
			}
			inserted++
			log.Printf("inserted: %s", imageURL)
		}

		if !res.IsTruncated {
			break
		}
		marker = res.NextMarker
	}

	log.Printf("Done. Inserted: %d, Skipped: %d", inserted, skipped)
}
