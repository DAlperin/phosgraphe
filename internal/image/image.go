package image

import (
	"fmt"
	"github.com/DAlperin/phosgraphe/internal/instructions"
	"github.com/DAlperin/phosgraphe/internal/utils"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"net/url"
	"time"
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
	var variantData []byte

	namespace := c.Params("namespace")
	unescapedNamespace, err := url.PathUnescape(namespace)
	if err != nil {
		return utils.SendError(c, err)
	}

	name := c.Params("name")
	unescapedName, err := url.PathUnescape(name)
	if err != nil {
		return utils.SendError(c, err)
	}

	if len(c.Params("*")) > 0 {
		parsed = instructions.Parse(c.Params("*"))
		hash = instructions.Hash(parsed)
	}

	image, isVariant, err := u.ImageService.Find(unescapedNamespace, unescapedName, hash)
	if image == nil {
		image, err := u.ImageService.FindBase(unescapedNamespace, unescapedName)
		if err != nil {
			return utils.SendError(c, err)
		}
		if image == nil {
			return utils.SendError(c, err)
		} else if len(parsed) > 0 {
			_, variantData, err = u.ImageService.BuildVariant(unescapedName, unescapedNamespace, parsed, hash)
			if err != nil {
				return utils.SendError(c, err)
			}
		}
	} else if err != nil {
		return utils.SendError(c, err)
	}

	var imgData []byte
	if variantData != nil {
		imgData = variantData
		err = nil
	} else if isVariant {
		imgData, err = u.ImageService.DownloadVariant(*image)
	} else {
		imgData, err = u.ImageService.Download(*image)
	}

	if err != nil {
		return utils.SendError(c, err)
	}

	contentType := http.DetectContentType(imgData)
	c.Set("Content-type", contentType)
	c.Set("Cache-Control", fmt.Sprintf("max-age=%d", 24*time.Hour*7))

	return c.Send(imgData)
}
