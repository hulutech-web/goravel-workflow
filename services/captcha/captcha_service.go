package captcha

import (
	"encoding/json"
	"fmt"
	"github.com/wenlng/go-captcha-assets/helper"
	"log"
	"strconv"

	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/rotate"
)

type CaptchaService struct {
}

var rotateCapt rotate.Captcha

func init() {
	builder := rotate.NewBuilder(rotate.WithRangeAnglePos([]option.RangeVal{
		{Min: 20, Max: 330},
	}))

	// background images
	imgs, err := images.GetImages()
	if err != nil {
		log.Fatalln(err)
	}

	// set resources
	builder.SetResources(
		rotate.WithImages(imgs),
	)

	rotateCapt = builder.Make()
}

func (c *CaptchaService) Generate() (string, int, string, string, error) {
	code := 0
	image_base64 := ""
	thumb_base64 := ""
	captchaData, err := rotateCapt.Generate()
	dotsByte, _ := json.Marshal(captchaData.GetData())
	captcha_key := helper.StringToMD5(string(dotsByte))
	//加入缓存
	WriteCache(captcha_key, dotsByte)
	if err != nil {
		return captcha_key, code, image_base64, thumb_base64, err
	}
	image_base64 = captchaData.GetMasterImage().ToBase64()
	thumb_base64 = captchaData.GetThumbImage().ToBase64()

	return captcha_key, code, image_base64, thumb_base64, nil

}

func (c *CaptchaService) CheckAngle(angle string, key string) (int, bool) {
	code := 1
	if angle == "" || key == "" {
		return code, false
	}

	cacheDataByte := ReadCache(key)
	if len(cacheDataByte) == 0 {
		return code, false
	}

	var dct *rotate.Block
	if err := json.Unmarshal(cacheDataByte, &dct); err != nil {
		return code, false
	}

	sAngle, _ := strconv.ParseFloat(fmt.Sprintf("%v", angle), 64)
	chkRet := rotate.CheckAngle(int64(sAngle), int64(dct.Angle), 2)

	if chkRet {
		code = 0
		return code, true
	} else {
		return code, false
	}
}
