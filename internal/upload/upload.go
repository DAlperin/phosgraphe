package upload

import (
	"github.com/DAlperin/phosgraphe/internal/image"
	"github.com/DAlperin/phosgraphe/internal/utils"
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
	ID                   string `form:"ID"`
	Namespace            string `form:"namespace"`
	EagerInstructions    string `form:"instructions"`
	IncomingInstructions string `form:"incomingInstructions"`
}

func (u *Handler) handleUpload(c *fiber.Ctx) error {
	formFile, err := c.FormFile("image")
	if err != nil {
		return utils.SendError(c, err)
	}

	var params uploadParams
	if err := c.BodyParser(&params); err != nil {
		return utils.SendError(c, err)
	}

	unescapedName, err := url.PathUnescape(params.ID)
	if err != nil {
		return utils.SendError(c, err)
	}

	unescapedNamespace, err := url.PathUnescape(params.Namespace)
	if err != nil {
		return utils.SendError(c, err)
	}

	if u.ImageService.Exists(unescapedName, unescapedNamespace) {
		return utils.SendError(c, err)
	}

	file, err := formFile.Open()
	if err != nil {
		return utils.SendError(c, err)
	}

	res, err := u.ImageService.Upload(file, unescapedNamespace, unescapedName)
	if err != nil {
		return utils.SendError(c, err)
	}

	err = u.ImageService.Store(res)
	if err != nil {
		return utils.SendError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "OK",
	})
}
