package image

import (
	"bytes"
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"github.com/DAlperin/phosgraphe/internal/models"
	"github.com/DAlperin/phosgraphe/internal/transforms"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type Service struct {
	DB               *gorm.DB
	S3Svc            *s3.S3
	Uploader         *s3manager.Uploader
	Downloader       *s3manager.Downloader
	TransformManager *transforms.TransformManager
}

func (s *Service) Exists(name string, namespace string) bool {
	var existing models.Image
	r := s.DB.Where("pretty_name = ? AND namespace = ?", name, namespace).First(&existing)
	if r.Error != nil {
		return false
	}
	return true
}

func (s *Service) Store(image *models.Image) error {
	r := s.DB.Create(image)
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
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	fmt.Println(contentType)
	up, err := s.Uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(viper.GetString("AWS_BUCKET")),
		Key:         aws.String(fmt.Sprintf("%s-%s", namespace, name)),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(up.Location)
	return &models.Image{
		PrettyName: name,
		Namespace:  namespace,
		Source:     up.Location,
	}, nil
}

func (s *Service) Find(namespace string, name string, hash string) (*models.Image, bool, error) {
	var image models.Image
	r := s.DB.Where("pretty_name = ? AND namespace = ? AND hash = ?", name, namespace, hash).First(&image)
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
	r := s.DB.Where("pretty_name = ? AND namespace = ?", name, namespace).First(&image)
	if r.Error != nil {
		return nil, r.Error
	}
	return &image, nil
}

func (s *Service) GetLink(image models.Image) (*string, error) {
	req, _ := s.S3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("AWS_BUCKET")),
		Key:    aws.String(fmt.Sprintf("%s-%s", image.Namespace, image.PrettyName)),
	})
	urlStr, err := req.Presign(1 * time.Minute)
	if err != nil {
		return nil, err
	}
	return &urlStr, nil
}

func (s *Service) GetVariantLink(image models.Image) (*string, error) {
	req, _ := s.S3Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("AWS_BUCKET")),
		Key:    aws.String(fmt.Sprintf("%s-%s-%s", image.Namespace, image.PrettyName, image.Hash)),
	})
	urlStr, err := req.Presign(1 * time.Minute)
	if err != nil {
		return nil, err
	}
	return &urlStr, nil
}

func (s *Service) BuildVariant(name string, namespace string, steps instructions.Instructions, hash string) (*models.Image, error) {
	buff := &aws.WriteAtBuffer{}
	_, _ = s.Downloader.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(viper.GetString("AWS_BUCKET")),
		Key:    aws.String(fmt.Sprintf("%s-%s", namespace, name)),
	})

	fmt.Println(len(buff.Bytes()))
	//err := os.WriteFile("./testing", buff.Bytes(), 0644)
	//fmt.Println(buff.Bytes())
	//fmt.Println(http.DetectContentType(buff.Bytes()))
	newVariant, err := s.applyTransform(buff.Bytes(), steps)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(newVariant)
	contentType := http.DetectContentType(newVariant)

	up, err := s.Uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(viper.GetString("AWS_BUCKET")),
		Key:         aws.String(fmt.Sprintf("%s-%s-%s", namespace, name, hash)),
		Body:        r,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}
	image := &models.Image{
		PrettyName: name,
		Namespace:  namespace,
		Source:     up.Location,
		Hash:       hash,
	}
	err = s.Store(image)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (s *Service) applyTransform(image []byte, steps instructions.Instructions) ([]byte, error) {
	buf := image
	for _, step := range steps {
		fmt.Println(step)
		for _, instruction := range step.Instructions {
			fmt.Println(instruction)
			handler, err := s.TransformManager.GetHandler(instruction)
			if err != nil {
				return nil, err
			}
			buf, err = handler.Transform(buf, instruction)
			if err != nil {
				return nil, err
			}
			fmt.Println(handler)
		}
	}
	return buf, nil
}
