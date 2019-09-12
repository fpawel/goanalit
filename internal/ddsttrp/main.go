package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/gobuffalo/packr"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

//go:generate packr

func main() {
	flag.StringVar(&targetFolderPath, "f", "", "target folder path")
	flag.BoolVar(&help, "?", false, "usage")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	if targetFolderPath == "" {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		targetFolderPath = dir
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	addr := "http://" + ln.Addr().String()
	fmt.Println(addr)

	boxAssets := packr.NewBox("./assets")
	fsAssets := http.StripPrefix("/assets/", http.FileServer(boxAssets))
	http.HandleFunc("/assets/", fsAssets.ServeHTTP)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		b, err := ioutil.ReadFile(filepath.Join(targetFolderPath, "errors.log"))
		if err != nil {
			panic(fmt.Sprintf("errors.log: %v", err))
		}
		mainPage := MainPage{}

		scanner := bufio.NewScanner(bytes.NewReader(b))
		for scanner.Scan() {
			s := strings.TrimSpace(scanner.Text())
			if len(s) == 0 {
				continue
			}
			xs := strings.Fields(s)
			if len(xs) < 3 {
				continue
			}
			createdAt, err := time.Parse("02.01.2006.15:04:05.000MST", xs[0])

			if err != nil {
				fmt.Println(xs, err)
				continue
			}

			mainPage.Records = append(mainPage.Records, ExceptionRecord{
				CreatedAt: createdAt,
				Class:     xs[1],
				Message:   strings.Join(xs[2:], " "),
				Ref:       stackTraceRef(createdAt),
			})

		}

		WriteMainHTML(w, &mainPage)
	})
	http.HandleFunc("/stack_trace/", func(w http.ResponseWriter, r *http.Request) {
		fileName := path.Base(r.RequestURI)

		file, err := os.Open(filepath.Join(targetFolderPath, "stack_trace", fileName))
		if err != nil {
			panic(err)
		}
		defer file.Close()

		var (
			p      StackTracePage
			lineNo int
		)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			switch lineNo {
			case 0:
				p.Class = scanner.Text()
			case 1:
				p.Message = scanner.Text()
			default:
				s := strings.TrimSpace(scanner.Text())
				if len(s) > 0 {
					line := strings.Split(scanner.Text(), "    ")
					if len(line) == 7 {
						p.StackTrace = append(p.StackTrace, line)
					}
				}
			}
			lineNo++
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		WriteMainHTML(w, &p)
	})

	go func() {

		if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", addr).Start(); err != nil {
			panic(err)
		}
	}()

	if err := http.Serve(ln, nil); err != nil {
		panic(err)
	}
}

var (
	targetFolderPath string

	help bool
)

func stackTraceRef(t time.Time) string {
	return t.Format("/stack_trace/02_01_2006_15_04_05_") + t.Format(".000")[1:] + ".stacktrace"
}
