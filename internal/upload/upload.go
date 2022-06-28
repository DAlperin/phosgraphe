package upload

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/image"
	"github.com/gofiber/fiber/v2"
	"net/url"
)

type Handler struct {
	ImageService image.Service
}

func (u *Handler) RegisterHandlers(r fiber.Router) {
	r.Post("/", u.handleUpload)
}

type uploadParams struct {
	ID                   string `form:"ID"'`
	Namespace            string `form:"namespace"`
	EagerInstructions    string `form:"instructions"`
	IncomingInstructions string `form:"incomingInstructions"`
}

func (u *Handler) handleUpload(c *fiber.Ctx) error {
	formFile, err := c.FormFile("image")
	if err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	//file, err := formFile.Open()
	//if err != nil {
	//	c.Status(500)
	//	return c.JSON(fiber.Map{
	//		"error": err.Error(),
	//	})
	//}
	//defer file.Close()

	var params uploadParams
	if err := c.BodyParser(&params); err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	unescapedName, err := url.PathUnescape(params.ID)
	if err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	unescapedNamespace, err := url.PathUnescape(params.Namespace)
	if err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if u.ImageService.Exists(unescapedName, unescapedNamespace) {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("'%s' already exists in namespace", params.ID),
		})
	}
	file, err := formFile.Open()
	if err != nil {
		return err
	}
	res, err := u.ImageService.Upload(file, unescapedNamespace, unescapedName)
	if err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("error uploading to s3: %s", err.Error()),
		})
	}

	err = u.ImageService.Store(res)
	if err != nil {
		c.Status(500)
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("error persisting to db: %s", err.Error()),
		})
	}

	return c.JSON(fiber.Map{
		"Hello": "world",
	})
}
