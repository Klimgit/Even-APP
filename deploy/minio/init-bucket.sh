#!/bin/sh
set -eu

until mc alias set local "$S3_ENDPOINT" "$S3_ACCESS_KEY" "$S3_SECRET_KEY"; do
  echo "waiting for minio..."
  sleep 2
done

mc mb --ignore-existing "local/${S3_BUCKET}"
echo "bucket ${S3_BUCKET} ready"
