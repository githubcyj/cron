package util

import (
	"bufio"
	"bytes"
	"github.com/crontab/master/manager"
	"os"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/20 18:58
  @todo
*/
//判断文件中是否包含特定字符
func IsFindStrInFile(str, path string) bool {
	var (
		f       *os.File
		err     error
		scanner *bufio.Scanner
		input   []byte
		sbyte   []byte
	)
	//打开文件
	if f, err = os.Open(path); err != nil {
		manager.GLogMgr.WriteLog("文件打开失败：" + err.Error())
	}
	defer f.Close()
	sbyte = []byte(str)
	scanner = bufio.NewScanner(f) //扫描文件
	for scanner.Scan() {          //扫描空格或者EOF,按行扫描
		input = scanner.Bytes()
		if bytes.Contains(input, sbyte) {
			manager.GLogMgr.WriteLog("文件中包含相同字符：" + string(input))
			return true
		}
	}

	return false
}
