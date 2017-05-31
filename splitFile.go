package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var target string

const (
	DEFAULT_PORT = "8080"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var clickedbutton = r.FormValue("clicked_button")
		newpath := filepath.Join(".", "/temp/")
		source := newpath + "/"
		removeContents(source)
		os.MkdirAll(newpath, os.ModePerm)
		if len(clickedbutton) > 0 {
			if strings.Compare(clickedbutton, "split") == 0 {
				fmt.Println("inside split file")
				file1, header1, err1 := r.FormFile("file1")
				if err1 != nil {
					panic(err1)
				}
				if file1 != nil {
					data1, dataerr1 := ioutil.ReadAll(file1)
					if dataerr1 != nil {
						panic(dataerr1)
					}
					var slice = r.FormValue("slice")
					splitfile(data1, header1.Filename, slice, source, newpath)
				}
			} else if strings.Compare(clickedbutton, "join") == 0 {
				fmt.Println("Inside join file")
				file2, header2, err2 := r.FormFile("file2")
				file3, header3, err3 := r.FormFile("file3")
				if file2 != nil && file3 != nil {
					data2, dataerr2 := ioutil.ReadAll(file2)
					data3, dataerr3 := ioutil.ReadAll(file3)
					if dataerr2 != nil {
						panic(dataerr2)
					}
					if dataerr3 != nil {
						panic(dataerr3)
					}
					var joinfilesArray = header2.Filename + "," + header3.Filename
					if len(joinfilesArray) > 0 {
						joinfile(joinfilesArray, data2, data3, source, newpath)
					}

				}

				if err2 != nil {
					panic(err2)
				}
				if err3 != nil {
					panic(err2)
				}

			}
		}
		var urlvalue = r.URL.Path[1:]
		if strings.Compare(urlvalue, "confirmation.html") == 0 {
			if len(target) > 0 {
				w.Header().Set("Content-Type", "applicaiton/zip")
				w.Header().Set("Content-Disposition", "attachment; filename=files.zip")
				http.ServeFile(w, r, target)
			}

		} else {
			http.ServeFile(w, r, r.URL.Path[1:])
		}

	})
	var port string
	if port = os.Getenv("PORT"); len(port) == 0 {
		port = DEFAULT_PORT
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func splitfile(x []byte, filename string, slice string, source string, newpath string) {
	fmt.Println("entering splitfile")

	if len(filename) > 0 {

		var extension = filepath.Ext(filename)
		fileNameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		var counter int
		slices, _ := strconv.Atoi(slice)
		sz := len(x)
		var sliceCount int
		sliceSize := sz / slices
		copyOfSliceSize := sz / slices

		target = source + "files.zip"
		//cleans the previous files.

		for counter < slices {
			counter++
			var buffer bytes.Buffer
			//buffer.WriteString(directory)
			buffer.WriteString(newpath)
			buffer.WriteString("/")
			buffer.WriteString(fileNameWithoutExt)
			buffer.WriteString("_")
			t := strconv.Itoa(counter)
			buffer.WriteString(t)
			buffer.WriteString(extension)

			_, err := os.Create(buffer.String())
			if err != nil {
				panic(err)
			}
			song := x[sliceCount:(sliceSize)]
			sliceCount = sliceSize
			sliceSize += copyOfSliceSize
			err = ioutil.WriteFile(buffer.String(), song, 0644)
			if err != nil {
				panic(err)
			}
		}

		zipit(source, target)
	}
}

func joinfile(fileArray string, x []byte, y []byte, source string, newpath string) {

	fmt.Println("entering joinfile")
	var files []string
	files = strings.Split(fileArray, ",")
	//check file types are same
	if checkfileType(files) {
		var z []byte
		z = x
		var j int
		for j = 0; j < len(y); j++ {
			z = append(z, y[j])
		}

		var buffer bytes.Buffer
		//buffer.WriteString(directory)
		buffer.WriteString(newpath)
		buffer.WriteString("/")
		buffer.WriteString("Merge")
		buffer.WriteString(filepath.Ext(files[0]))
		_, err := os.Create(buffer.String())
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(buffer.String(), z, 0644)
		if err != nil {
			panic(err)
		}
		target = source + "files.zip"
		//create the zipfile
		zipit(source, target)

	} else {
		fmt.Println("No call")
	}
}

func checkfileType(files []string) bool {
	fmt.Println("entering checkfileType")
	var i int
	var result bool
	for i = 0; i < len(files); i++ {
		var extension = filepath.Ext(files[i])
		if i == 0 {
			continue
		} else {
			result = strings.HasSuffix(files[i-1], extension)
			if result == false {
				break
			}
		}
	}

	return result
}

func zipit(source, target string) error {
	fileList := []string{}
	filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	sz := len(fileList)
	fileList = fileList[1:sz]

	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, file := range fileList {

			zipFile, err := os.Open(file)
			if err != nil {
				return err
			}
			defer zipFile.Close()

			// get the file information
			info, err := zipFile.Stat()
			if err != nil {
				return err
			}

			// create a zip header from the os.FileInfo
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// write the header to the zip file
			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			// copy the file to the zip file
			_, err = io.Copy(writer, zipFile)
			if err != nil {
				return err
			}
			os.Remove(file)
		}
		return err
	})
	return nil
}

//remove the content of folder
func removeContents(dir string) {
	if _, err := os.Stat(dir); err == nil {
		d, _ := os.Open(dir)
		fmt.Println(dir)
		os.RemoveAll(dir)
		os.RemoveAll("/temp/")
		defer d.Close()
	}

}
