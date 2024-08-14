package Upload

import (
	"fmt"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
	"time"
)

type UploadService struct {
}
type UploadImport struct {
	Path   string `json:"path" form:"path"`   //附件路径
	Alias  string `json:"alias" form:"alias"` //自定义表名
	Id     int    `json:"id" form:"id"`       //附件id attachment_id
	UserId int    `json:"user_id" form:"user_id"`
}

func NewUploadService() *UploadService {
	return &UploadService{}
}
func (*UploadService) Upload(ctx http.Context, file filesystem.File) (*models.Attachment, error) {
	name := file.GetClientOriginalName()
	extension := file.GetClientOriginalExtension()
	//获取当前年月2022-02
	yearMonth := fmt.Sprintf("%d-%02d", time.Now().Year(), time.Now().Month())
	//file, err := ctx.Request().File(name)
	putFile, err := facades.Storage().PutFile(yearMonth, file)

	//保存文件，返回保存路径
	user := models.User{}
	if err1 := facades.Auth(ctx).User(&user); err1 != nil {
		att := models.Attachment{
			Name: name,
			Path: putFile,
			Ext:  extension,
		}
		err1 := facades.Orm().Query().Model(&models.Attachment{}).Create(&att)
		if err1 != nil {
			return nil, err1
		}
		return &att, nil
	}
	att := models.Attachment{
		Name:   name,
		Path:   putFile,
		UserID: user.ID,
		Ext:    extension,
	}
	err2 := facades.Orm().Query().Model(&models.Attachment{}).Create(&att)
	if err != nil {
		return nil, err2
	}
	return &att, nil
}
