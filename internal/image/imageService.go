package image

import (
	"bytes"
	"fmt"
	"github.com/DAlperin/phosgraphe/internal"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"github.com/DAlperin/phosgraphe/internal/models"
	"github.com/DAlperin/phosgraphe/internal/transforms"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Service struct {
	db               *gorm.DB
	s3Svc            *s3.S3
	uploader         *s3manager.Uploader
	downloader       *s3manager.Downloader
	transformManager *transforms.TransformManager
	config           internal.Config
}

func NewService(db *gorm.DB, s3svc *s3.S3, uploader *s3manager.Uploader, downloader *s3manager.Downloader, transformManager *transforms.TransformManager, config internal.Config) *Service {
	return &Service{
		db:               db,
		s3Svc:            s3svc,
		uploader:         uploader,
		downloader:       downloader,
		transformManager: transformManager,
		config:           config,
	}
}

func (s *Service) Exists(name string, namespace string) bool {
	var existing models.Image
	r := s.db.Where("pretty_name = ? AND namespace = ?", name, namespace).First(&existing)
	if r.Error != nil {
		return false
	}
	return true
}

func (s *Service) Store(image *models.Image) error {
	r := s.db.Create(image)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func GetFileContentType(out multipart.File) (string, error) {
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func (s *Service) Upload(file multipart.File, namespace string, name string) (*models.Image, error) {
	contentType, err := GetFileContentType(file)
	if err != nil {
		return nil, err
	}
	//Don't accidentally truncate the image. The image will still *technically* be readable but ImageMagick will be mad
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	up, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.config.AwsBucket),
		Key:         aws.String(fmt.Sprintf("%s-%s", namespace, name)),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return &models.Image{
		PrettyName: name,
		Namespace:  namespace,
		Source:     up.Location,
	}, nil
}

func (s *Service) Find(namespace string, name string, hash string) (*models.Image, bool, error) {
	var image models.Image
	r := s.db.Where("pretty_name = ? AND namespace = ? AND hash = ?", name, namespace, hash).First(&image)
	if r.Error != nil {
		return nil, false, r.Error
	}
	if len(hash) > 0 {
		return &image, true, nil
	}
	return &image, false, nil
}

func (s *Service) FindBase(namespace string, name string) (*models.Image, error) {
	var image models.Image
	r := s.db.Where("pretty_name = ? AND namespace = ?", name, namespace).First(&image)
	if r.Error != nil {
		return nil, r.Error
	}
	return &image, nil
}

// FIXME: we need to find a way to cache the pre-signed links

func (s *Service) GetLink(image models.Image) (*string, error) {
	req, _ := s.s3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.config.AwsBucket),
		Key:    aws.String(fmt.Sprintf("%s-%s", image.Namespace, image.PrettyName)),
	})
	urlStr, err := req.Presign(20 * time.Minute)
	if err != nil {
		return nil, err
	}
	return &urlStr, nil
}

func (s *Service) GetVariantLink(image models.Image) (*string, error) {
	req, _ := s.s3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket:               aws.String(s.config.AwsBucket),
		Key:                  aws.String(fmt.Sprintf("%s-%s-%s", image.Namespace, image.PrettyName, image.Hash)),
		ResponseCacheControl: aws.String(fmt.Sprintf("max-age=%d", 24*time.Hour*7)),
	})
	urlStr, err := req.Presign(20 * time.Minute)
	if err != nil {
		return nil, err
	}
	return &urlStr, nil
}

func (s *Service) download(key string) ([]byte, error) {
	buff := &aws.WriteAtBuffer{}
	_, err := s.downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(s.config.AwsBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil

}

func (s *Service) DownloadVariant(image models.Image) ([]byte, error) {
	return s.download(fmt.Sprintf("%s-%s-%s", image.Namespace, image.PrettyName, image.Hash))
}

func (s *Service) Download(image models.Image) ([]byte, error) {
	return s.download(fmt.Sprintf("%s-%s", image.Namespace, image.PrettyName))
}

func (s *Service) BuildVariant(name string, namespace string, steps instructions.Instructions, hash string) (*models.Image, []byte, error) {
	buff := &aws.WriteAtBuffer{}
	_, _ = s.downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(s.config.AwsBucket),
		Key:    aws.String(fmt.Sprintf("%s-%s", namespace, name)),
	})

	newVariant, err := s.applyTransform(buff.Bytes(), steps)
	if err != nil {
		return nil, nil, err
	}

	r := bytes.NewReader(newVariant)
	contentType := http.DetectContentType(newVariant)

	up, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.config.AwsBucket),
		Key:         aws.String(fmt.Sprintf("%s-%s-%s", namespace, name, hash)),
		Body:        r,
		ContentType: aws.String(contentType),
		Metadata: map[string]*string{
			"transform": aws.String(steps.String()),
		},
	})
	if err != nil {
		return nil, nil, err
	}
	image := &models.Image{
		PrettyName: name,
		Namespace:  namespace,
		Source:     up.Location,
		Hash:       hash,
	}
	err = s.Store(image)
	if err != nil {
		return nil, nil, err
	}
	return image, newVariant, nil
}

func (s *Service) applyTransform(image []byte, steps instructions.Instructions) ([]byte, error) {
	var err error

	buf := image
	handlers := map[string]transforms.Transformation{}

	for _, step := range steps {
		for _, instruction := range step.Instructions {
			handler, err := s.transformManager.GetHandler(instruction)
			if err != nil {
				return nil, err
			}
			handlers[instruction] = handler
		}
	}

	for instruction, handler := range handlers {
		buf, err = handler.Transform(buf, instruction)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}
