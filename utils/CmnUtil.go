package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type cmnUtil struct{}

var CmnUtil = new(cmnUtil)

// 获得服务器根网址
//
// @param `str` `xx/xx/`
//
// @return `http://xx:xx/`
func (c *cmnUtil) GetServerUrl(str string) string {
	return fmt.Sprintf("http://127.0.0.1:1234/%s", str)
}

// 字符串首字母大写
func (c *cmnUtil) Capitalize(str string) string {
	return strings.Replace(str, string(str[0]), strings.ToUpper(string(str[0])), 1)
}

func (c *cmnUtil) CurrentTimeInt() string {
	return time.Now().Format("20060102150405")
}

// 判断是否是受支持图片格式
// jpg png jpeg bmp
//
// @return `ok` `ext`
func (c *cmnUtil) ImageExt(filename string) (bool, string) {
	ext := path.Ext(filename)
	return ext == ".jpg" ||
			ext == ".png" ||
			ext == ".jpeg" ||
			ext == ".bmp",
		ext
}

// 判断是否是受支持视频格式
// mp4
//
// @return `ok` `ext`
func (c *cmnUtil) VideoExt(filename string) (bool, string) {
	ext := path.Ext(filename)
	return ext == ".mp4",
		ext
}

// 判断文件或文件夹是否存在
func (c *cmnUtil) IsFileExist(filename string) bool {
	_, err := os.Stat(filename)
	// os.IsNotExist(err)
	return err == nil
}

// 保存文件，并且覆盖已存在文件
func (c *cmnUtil) SaveFile(filename string, file io.Reader) bool {

	// File path
	dir := path.Dir(filename)
	if !c.IsFileExist(dir) {
		os.MkdirAll(dir, 0777)
		if !c.IsFileExist(dir) {
			return false
		}
	}

	// Delete file
	if c.IsFileExist(filename) {
		err := os.Remove(filename)
		if err != nil {
			return false
		}
	}

	// Create file
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return false
	}

	// File exist
	if !c.IsFileExist(filename) {
		return false
	}

	// Save file
	_, err = io.Copy(f, file)
	if err != nil {
		return false
	}

	return true
}
