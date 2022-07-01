package utils

import "github.com/gofiber/fiber/v2"

func SendError(c *fiber.Ctx, err error) error {
	c.Status(500)
	return c.JSON(fiber.Map{
		"error": err.Error(),
	})
}
