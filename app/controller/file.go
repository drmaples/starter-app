package controller

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/drmaples/starter-app/app/dto"
	"github.com/drmaples/starter-app/app/objectstore"
)

func handleListFiles(c echo.Context) error {
	ctx := c.Request().Context()

	// FIXME: do not construct this every time
	bucket, ok := os.LookupEnv("OBJECT_STORE_BUCKET")
	if !ok {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp("object store bucket not defined"))
	}

	objStore, err := objectstore.New(ctx, bucket)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}
	obj, err := objStore.GetObject(ctx, "darrell.txt")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResp(err.Error()))
	}

	return c.JSON(http.StatusOK, dto.FileResponse{Key: obj.Key, Data: string(obj.Data)})
}
