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

func (c *LogWatcher) GetFileList(newFileMap map[string]*os.File) map[int]*os.File {
	/*
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
		fileMap := c.GetFileList(newFileMap)

		for _, v := range fileMap {
			// var totalStr string
			var totalSize int32 = 0

			buf := make([]byte, 1024)
			var isFirst = 1
			for {

				n, _ := v.Read(buf)
				if 0 == n {
					break
				}
				totalSize = totalSize + int32(n)
				if totalSize > 0 {
					if isFirst == 1 {
						isFirst = 0
						fmt.Printf("==================File:%s=================\n", v.Name())
					}

					fmt.Print(string(buf[:n]))

				}
				//totalStr = totalStr + string(buf[:n])
				// fmt.Print(string(buf[:n]))
			}
			// if(totalSize > 0){

			// 	fmt.Print(totalStr)
			// }
		}

		time.Sleep(time.Millisecond * 10)

	}

	return
}
