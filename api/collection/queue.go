package collect

import "go.uber.org/zap"

func (c *CollectAPI) pushDeleteJPGQueue(filename string) {
	_, err := c.Storage.DeleteJPG(filename)
	if err != nil {
		logger.Error("delete jpg", zap.String("error", err.Error()))
	}
}
