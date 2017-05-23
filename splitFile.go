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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var clickedbutton = r.FormValue("clicked_button")

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
					splitfile(data1, header1.Filename, slice)
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
						joinfile(joinfilesArray, data2, data3)
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

		fmt.Printf("url value %s", r.URL.Path[1:])
		var urlvalue = r.URL.Path[1:]
		if strings.Compare(urlvalue, "confirmation.html") == 0 {
			fmt.Println("inside urlvalue")
			w.Header().Set("Content-Type", "applicaiton/zip")
			w.Header().Set("Content-Disposition", "attachment; filename=files.zip")
			target := "\\temp" + "\\files.zip"
			http.ServeFile(w, r, target)
		} else {
			http.ServeFile(w, r, r.URL.Path[1:])
		}

	})
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func splitfile(x []byte, filename string, slice string) {
	fmt.Println("entering splitfile")

	if len(filename) > 0 {
		var extension = filepath.Ext(filename)
		var directory = filepath.Dir(filename)
		fileNameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		var counter int
		slices, _ := strconv.Atoi(slice)
		sz := len(x)
		fmt.Println(sz)
		var sliceCount int
		sliceSize := sz / slices
		copyOfSliceSize := sz / slices
		fmt.Println(sliceSize)
		for counter < slices {
			counter++
			var buffer bytes.Buffer
			buffer.WriteString(directory)
			buffer.WriteString("/temp/")
			buffer.WriteString(fileNameWithoutExt)
			buffer.WriteString("_")
			t := strconv.Itoa(counter)
			buffer.WriteString(t)
			buffer.WriteString(extension)
			os.Create(buffer.String())
			song := x[sliceCount:(sliceSize)]
			sliceCount = sliceSize
			sliceSize += copyOfSliceSize
			var err = ioutil.WriteFile(buffer.String(), song, 0644)
			if err != nil {
				panic(err)
			}
			target := directory + "\\temp" + "\\files.zip"
			fmt.Printf("target:%s", target)
			zipit(buffer.String(), target)
		}
	}
}

func joinfile(fileArray string, x []byte, y []byte) {

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
		var mergefileName = filepath.Dir(files[0]) + "\\temp" + "\\Merge" + filepath.Ext(files[0])
		os.Create(mergefileName)
		err := ioutil.WriteFile(mergefileName, z, 0644)
		if err != nil {
			panic(err)
		}
		target := filepath.Dir(files[0]) + "\\temp" + "\\files.zip"
		//create the zipfile
		zipit(mergefileName, target)

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
	zipfile, err := os.Create(target)

	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)

	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	fmt.Println(info)

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
