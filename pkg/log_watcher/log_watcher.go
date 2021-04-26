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
		Pattern:   FormatPattern(pattern),
		FileMap:   make(map[string]*os.File),
		IsStopped: 0,
	}
	return obj

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

func (c *LogWatcher) ReplaceWith(newFileMap map[string]*os.File) {
	// 替换
	err := func() error {
		/*
			1. 如果在old里面没有找到，就添加
			2. 如果在old里面找到了
				a. 如果sameFile, 关闭当前的，然后退出
				b. 关闭old, 然后添加
		*/
		for name, newFile := range newFileMap {
			oldFile, exists := c.FileMap[name]
			if !exists {
				// add
				c.FileMap[name] = newFile
				continue
			}
			// exists
			oldFileStat, err := os.Stat(oldFile.Name())
			if err != nil {
				return err
			}
			newFildStat, err := os.Stat(newFile.Name())
			if err != nil {
				return err
			}

			// 如果文件是一样的，关闭新的Fild
			if os.SameFile(oldFileStat, newFildStat) {
				err = newFile.Close()
				if err != nil {
					return err
				}
				continue
			}

			// if oldFile.Fd() == newFile.Fd() {
			// 	continue
			// }
			// oldFile.Fd != newFile.Fd()
			// 文件不一样，就表示是重建了文件
			err = oldFile.Close()
			if err != nil {
				return err
			}
			c.FileMap[name] = newFile
		}
		return nil
	}()
	if err != nil {
		panic(err)
	}
}

func (c *LogWatcher) ReplaceFileMap(newFileMap map[string]*os.File) {
	// 替换本地的FileMap
	err := func() error {

		oldFileMap := c.FileMap
		c.FileMap = newFileMap
		for fileName, oldFile := range oldFileMap {
			newFile, exists := newFileMap[fileName]
			if !exists {
				// 新文件列表里面没有
				err := oldFile.Close()
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
				err := newFile.Close()
				if err != nil {
					return err
				}
				c.FileMap[fileName] = oldFile
				continue
			}
			// 文件不一样
			err = oldFile.Close()
			if err != nil {
				return err
			}

		}

		return nil
	}()
	if err != nil {
		// log.Panic(err)
		panic(err)
		// log.Panicf("ReplaceFileMap Error %v", err)
	}

}

func (c *LogWatcher) GetFileList(newFileMap map[string]*os.File) map[int]*os.File {
	/*
		    获取最终要遍历的文件
			1. 添加新的
			2. 移除旧的东西
	*/
	var fileMap = make(map[int]*os.File)
	var fileIndex = 0

	err := func() error {
		for name, newFile := range newFileMap {
			oldFile, exists := c.FileMap[name]
			if !exists {
				// add
				// c.FileMap[name] = newFile
				fileMap[fileIndex] = newFile
				fileIndex++
				continue
			}
			// exists
			oldFileStat, err := os.Stat(oldFile.Name())
			if err != nil {
				return err
			}
			newFildStat, err := os.Stat(newFile.Name())
			if err != nil {
				return err
			}

			if os.SameFile(oldFileStat, newFildStat) {
				fileMap[fileIndex] = oldFile
				fileIndex++
				continue
			}

			// if oldFile.Fd() == newFile.Fd() {
			// 	continue
			// }
			// oldFile.Fd != newFile.Fd()
			//oldFile.Close()
			//c.FileMap[name] = newFile
			fileMap[fileIndex] = newFile
			fileIndex++

		}
		return nil
	}()

	if err != nil {
		panic(err)
	}
	return fileMap

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
		time.Sleep(time.Millisecond * 10) // sleep10毫秒

	}

	return
}
