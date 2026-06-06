package config

// S3 holds MinIO / S3-compatible storage settings.
type S3 struct {
	Endpoint       string
	PublicEndpoint string
	Bucket         string
	AccessKey      string
	SecretKey      string
}

// LoadS3 reads S3_* env vars. All fields are required for services that use media.
func LoadS3() (S3, error) {
	endpoint, err := MustGetenv("S3_ENDPOINT")
	if err != nil {
		return S3{}, err
	}
	public, err := MustGetenv("S3_PUBLIC_ENDPOINT")
	if err != nil {
		return S3{}, err
	}
	bucket, err := MustGetenv("S3_BUCKET")
	if err != nil {
		return S3{}, err
	}
	access, err := MustGetenv("S3_ACCESS_KEY")
	if err != nil {
		return S3{}, err
	}
	secret, err := MustGetenv("S3_SECRET_KEY")
	if err != nil {
		return S3{}, err
	}
	return S3{
		Endpoint:       endpoint,
		PublicEndpoint: public,
		Bucket:         bucket,
		AccessKey:      access,
		SecretKey:      secret,
	}, nil
}
