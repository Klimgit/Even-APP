#!/bin/sh
set -eu

until mc alias set local "$S3_ENDPOINT" "$S3_ACCESS_KEY" "$S3_SECRET_KEY"; do
  echo "waiting for minio..."
  sleep 2
done

mc mb --ignore-existing "local/${S3_BUCKET}"

# Browser uploads from Flutter web PUT directly to presigned MinIO URLs.
cat >/tmp/even-media-cors.xml <<'EOF'
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>http://localhost:5173</AllowedOrigin>
    <AllowedOrigin>http://127.0.0.1:5173</AllowedOrigin>
    <AllowedOrigin>http://localhost:8080</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
    <AllowedMethod>PUT</AllowedMethod>
    <AllowedMethod>HEAD</AllowedMethod>
    <AllowedMethod>POST</AllowedMethod>
    <AllowedHeader>*</AllowedHeader>
    <ExposeHeader>ETag</ExposeHeader>
    <MaxAgeSeconds>3600</MaxAgeSeconds>
  </CORSRule>
</CORSConfiguration>
EOF
mc cors set "local/${S3_BUCKET}" /tmp/even-media-cors.xml 2>/dev/null || \
  mc admin config set local api cors_allow_origin="http://localhost:5173,http://127.0.0.1:5173" 2>/dev/null || true

echo "bucket ${S3_BUCKET} ready"
