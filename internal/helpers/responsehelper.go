package helpers

import "github.com/gin-gonic/gin"

func ResError(c *gin.Context, code int, message interface{}, data interface{}) {
	c.JSON(code, gin.H{
		"message": message,
		"details": data,
	})

}
func Ok(c *gin.Context, code int, message interface{}, data interface{}) {
	c.JSON(code, gin.H{
		"message": message,
		"data":    data,
	})

}
