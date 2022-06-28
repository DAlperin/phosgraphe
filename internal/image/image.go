package image

import (
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"github.com/DAlperin/phosgraphe/internal/models"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/url"
)

type Handler struct {
	ImageService Service
}

func (u *Handler) RegisterHandlers(r fiber.Router) {
	r.Get("/:namespace/:name/*", u.GetImage)
}

func (u *Handler) GetImage(c *fiber.Ctx) error {
	var hash string
	var parsed instructions.Instructions
	var imageLink *string
	var variant *models.Image

	namespace := c.Params("namespace")
	unescapedNamespace, err := url.PathUnescape(namespace)
	if err != nil {
		log.Fatal(err)
	}

	name := c.Params("name")
	unescapedName, err := url.PathUnescape(name)
	if err != nil {
		log.Fatal(err)
	}

	if len(c.Params("*")) > 0 {
		parsed = instructions.Parse(c.Params("*"))
		hash = instructions.Hash(parsed)
	}

	image, isVariant, err := u.ImageService.Find(unescapedNamespace, unescapedName, hash)
	if image == nil {
		image, err := u.ImageService.FindBase(unescapedNamespace, unescapedName)
		if image == nil {
			return sendError(c, err)
		} else if len(parsed) > 0 {
			variant, err = u.ImageService.BuildVariant(unescapedName, unescapedNamespace, parsed, hash)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return sendError(c, err)
		}
	} else if err != nil {
		return sendError(c, err)
	}

	if isVariant {
		imageLink, err = u.ImageService.GetVariantLink(*image)
	} else if variant != nil {
		imageLink, err = u.ImageService.GetVariantLink(*variant)
	} else {
		imageLink, err = u.ImageService.GetLink(*image)
	}
	if err != nil {
		return sendError(c, err)
	}

	//If we get here we should have an image link either to a base or variant
	return c.Redirect(*imageLink, 302)
}

func sendError(c *fiber.Ctx, err error) error {
	c.Status(500)
	return c.JSON(fiber.Map{
		"error": err.Error(),
	})
}
