package mobile_captcha

import (
	"bytes"
	"math/rand"
	"strconv"
	"time"
)

// 验证码生成器
type CaptchaCodeGenerator interface {
	// 生成验证码
	Generate() string
}

// 基于math/rand的纯数字验证码生成器
type RandDigitalCaptchaCodeGenerator struct {
	rnd    *rand.Rand
	length int
}

// 创建一个基于math/rand的纯数字验证码生成器
func NewRandDigitalCaptchaGenerator(len int) *RandDigitalCaptchaCodeGenerator {
	return &RandDigitalCaptchaCodeGenerator{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
		length: len,
	}
}

func (gen *RandDigitalCaptchaCodeGenerator) Generate() string {
	buf := bytes.NewBufferString("")
	for i := 0; i < gen.length; i++ {
		buf.WriteString(strconv.Itoa(gen.rnd.Intn(10)))
	}
	return buf.String()
}
