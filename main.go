package main

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal"
	"github.com/DAlperin/phosgraphe/internal/image"
	"github.com/DAlperin/phosgraphe/internal/models"
	"github.com/DAlperin/phosgraphe/internal/transforms"
	"github.com/DAlperin/phosgraphe/internal/upload"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/spf13/viper"
	"gopkg.in/gographics/imagick.v3/imagick"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read in config: %s", err.Error())
	}

	var config internal.Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Failed to read in config: %s", err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		log.Fatalf(err.Error())
	}

	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", config.DbHost, config.DbUser, config.DbPass, config.DbName)), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err.Error())
	}

	if err = db.AutoMigrate(models.Image{}); err != nil {
		log.Fatalf("Failed to migrate database: %s", err.Error())
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(config.AwsID, config.AwsSecret, ""),
	}))

	s3Svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	app := fiber.New()
	app.Use(logger.New())

	uploadRoutes := app.Group("/upload")
	imageRoutes := app.Group("/image")

	imagick.Initialize()
	defer imagick.Terminate()

	transformManager := transforms.New()

	imageService := image.NewService(db, s3Svc, uploader, downloader, transformManager, config)

	uploadHandler := upload.Handler{
		ImageService: *imageService,
	}
	uploadHandler.RegisterHandlers(uploadRoutes)

	imageHandler := image.Handler{
		ImageService: *imageService,
	}
	imageHandler.RegisterHandlers(imageRoutes)

	if err = app.Listen(":8080"); err != nil {
		log.Fatal("Failed to start server")
	}
}
