package log_watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

/*
先暂时不实现 -n 10 这个参数
*/
type LogWatcher struct {
	Pattern   string
	FileMap   map[string]*os.File
	IsStopped int16
}

func NewLogWatcher(pattern string) *LogWatcher {
	var obj = &LogWatcher{
		Pattern:   pattern,
		FileMap:   make(map[string]*os.File),
		IsStopped: 0,
	}
	return obj

}

func (c *LogWatcher) Prepare() {

	c.FileMap = c.getInfo(c.Pattern)
}

func (c *LogWatcher) Compare(newFileMap map[string]*os.File) {
	/*
		1. 添加新的
		2. 移除旧的东西
	*/
	for name, newFile := range newFileMap {
		oldFile, exists := c.FileMap[name]
		if !exists {
			// add
			c.FileMap[name] = newFile
			continue
		}
		// exists

		if oldFile.Fd() == newFile.Fd() {
			continue
		}
		// oldFile.Fd != newFile.Fd()
		oldFile.Close()
		c.FileMap[name] = newFile

	}

}

func (c *LogWatcher) getInfo(pattern string) map[string]*os.File {
	var fileMap = make(map[string]*os.File)

	matches, err := filepath.Glob(c.Pattern)
	if err != nil {
		log.Fatal(err)
		return fileMap
	}
	for _, v := range matches {
		fileObj, err := os.Open(v)
		if err != nil {
			log.Fatal(err)
			return fileMap
		}
		// 移动到最后
		fileObj.Seek(0, 2)
		fileMap[v] = fileObj
	}

	return fileMap
}

func (c *LogWatcher) Tail() {

	for {
		if c.IsStopped == 1 {
			return
		}
		var newFileMap = c.getInfo(c.Pattern)
		c.Compare(newFileMap)

		for fileName, v := range c.FileMap {
			var totalStr string
			var totalSize int32 = 0
			

			for {
				buf := make([]byte, 1024)
				n, _ := v.Read(buf)
				if 0 == n {
					break
				}
				totalSize = totalSize + int32(n)
				totalStr = totalStr + string(buf[:n])
				// fmt.Print(string(buf[:n]))
			}
			if(totalSize > 0){
				fmt.Printf("==================File:%s=================\n" ,fileName)
				fmt.Print(totalStr)
			}
		}

		time.Sleep(time.Microsecond * 10)

	}

	return
}
