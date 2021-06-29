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
	Verbose   int
}

func NewLogWatcher(pattern string, verbose int) *LogWatcher {
	var obj = &LogWatcher{
		Pattern:   FormatPattern(pattern),
		FileMap:   make(map[string]*os.File),
		IsStopped: 0,
		Verbose:   verbose,
	}
	return obj

}

func (c LogWatcher) LogMessage(format string, v ...interface{}) {
	if c.Verbose == 1 {
		log.Printf(format, v...)
	}
}

func FormatPattern(pattern string) string {
	first := pattern[0]
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fail to getwd")
	}
	if first == '.' {
		return cwd + pattern[1:]
	}
	if first == '~' {
		return os.Getenv("HOME") + pattern[1:]
	}
	return pattern
}

func (c *LogWatcher) Prepare() {

	c.FileMap = c.getInfo(c.Pattern)
}

func (c *LogWatcher) ReplaceFileMap(newFileMap map[string]*os.File) {
	// 替换本地的FileMap
	err := func() error {

		oldFileMap := make(map[string]*os.File)
		for fileName, fileInfo := range c.FileMap {
			oldFileMap[fileName] = fileInfo
		}
		// c.FileMap = newFileMap
		c.FileMap = make(map[string]*os.File)
		for fileName, fileInfo := range newFileMap {
			c.FileMap[fileName] = fileInfo
		}
		for fileName, oldFile := range oldFileMap {
			newFile, exists := newFileMap[fileName]
			c.LogMessage("for file %s", fileName)
			if !exists {
				// 新文件列表里面没有
				err := oldFile.Close()
				c.LogMessage("file %s does not exists", fileName)
				if err != nil {
					return err
				}
				continue
			}

			oldFileStat, err := os.Stat(oldFile.Name())
			if err != nil {
				return err
			}
			newFildStat, err := os.Stat(newFile.Name())
			if err != nil {
				return err
			}

			if os.SameFile(oldFileStat, newFildStat) {
				// 文件一样
				// fileMap[fileIndex] = oldFile
				// fileIndex++
				// continue
				c.LogMessage("%s is the same", fileName)
				c.LogMessage(fmt.Sprintf("oldFileSys:%+v", oldFileStat.Sys()))
				c.LogMessage(fmt.Sprintf("NewFileSys:%+v", oldFileStat.Sys()))
				err := newFile.Close()
				if err != nil {
					return err
				}
				c.FileMap[fileName] = oldFile
				continue
			}
			// 文件不一样
			// 按理来讲不需要设置的呀
			// c.FileMap[fileName] = newFile
			// 为什么没有日志呢?
			// log.Printf("")
			c.LogMessage("%s file recreate", fileName)
			err = oldFile.Close()
			if err != nil {
				return err
			}
			// c.FileMap[file
			newFile.Seek(0, 0)

		}

		return nil
	}()
	if err != nil {
		// log.Panic(err)
		panic(err)
		// log.Panicf("ReplaceFileMap Error %v", err)
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
		// fmt.Printf("pattern:%s", c.Pattern)
		// fileMap := c.GetFileList(newFileMap)
		c.ReplaceFileMap(newFileMap)
		for name, _ := range newFileMap {
			// log.Printf("name:%s", name)
			c.LogMessage("name:%s", name)
		}

		for _, v := range c.FileMap {
			// var totalStr string
			var totalSize int32 = 0

			buf := make([]byte, 1024)
			var isFirst = 1
			for {

				n, _ := v.Read(buf)
				if 0 == n {
					break
				}
				// io.copy
				totalSize = totalSize + int32(n)
				if totalSize > 0 {
					if isFirst == 1 {
						isFirst = 0
						fmt.Printf("==================File:%s=================\n", v.Name())
						// os.Stdout.WriteString(fmt.Sprintf(""))
					}

					fmt.Print(string(buf[:n]))
					// io.Co

				}
				//totalStr = totalStr + string(buf[:n])
				// fmt.Print(string(buf[:n]))
			}
			// if(totalSize > 0){

			// 	fmt.Print(totalStr)
			// }
		}
		// 设置缓存的东西

		// 为什么这个程序偶尔，cpu会比较高，高达30-40%
		time.Sleep(time.Millisecond * 100) // sleep10毫秒

	}

	return
}
