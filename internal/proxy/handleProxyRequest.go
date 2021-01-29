package proxy

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nortoneo/iptv-proxy/internal/config"
	"github.com/nortoneo/iptv-proxy/internal/urlconvert"
)

// https://gist.github.com/gruber/249502
const urlRegex = `(?i)\b((?:[a-z][\w-]+:(?:/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s` + "`" + `!()\[\]{};:'".,<>?«»“”‘’]))`

func handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	currentURLString := r.URL.String()
	realURLString, listName, err := urlconvert.ConvertProxyURLtoURL(currentURLString)
	if err != nil {
		log.Println("Failed to convert current url: " + currentURLString)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	req, err := http.NewRequest("GET", realURLString, nil)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", r.Header.Get("user-agent"))
	resp, err := GetClient().Do(req)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("content-type")
	if contentType != "" {
		w.Header().Set("content-type", contentType)
	}

	location := resp.Header.Get("location")
	if location != "" {
		proxyLocation, err := urlconvert.ConvertURLtoProxyURL(location, config.GetConfig().AppURL, listName)
		if err != nil {
			log.Println("Unable to convert location header: " + location)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("Redirecting to: " + proxyLocation + " original target: " + location)
		w.Header().Set("location", proxyLocation)
	}

	w.Header().Set("X-Robots-Tag", "noindex, nofollow, nosnippet")
	w.WriteHeader(resp.StatusCode)

	//	handling body - decide by content type if we should stream the response or parse it to convert potential urls
	parsableContentType := [...]string{"text/", "url"}
	for _, parsableCT := range parsableContentType {
		if strings.Contains(contentType, parsableCT) {
			log.Println("Parsing: [" + contentType + "] " + realURLString)
			parseHTTPClientResponceBody(resp, w, listName)
			log.Println("Completed:  [" + contentType + "] " + realURLString)
			return
		}
	}

	streamableContentType := [...]string{"video/", "image/", "application/octet-stream"}
	for _, streamableCT := range streamableContentType {
		if strings.Contains(contentType, streamableCT) {
			log.Println("Streaming:  [" + contentType + "] " + realURLString)
			streamHTTPClientResponceBody(resp, w)
			log.Println("Completed:  [" + contentType + "] " + realURLString)
			return
		}
	}

	//content type not recognized, decide by file extension if we should stream or parse text
	pathExtension := ""
	realURL, err := url.Parse(realURLString)
	if err == nil {
		pathExtension = filepath.Ext(realURL.Path)
	}

	streamableFileExtension := [...]string{"ts", "h264", "mkv", "mpg", "mpeg", "mp2", "mpe", "mpv", "vob", "mp4", "m4p", "m4v", "avi", "mp3", "aac", "mpa", "ac3", "webm", "ogg", "mov"}
	for _, ext := range streamableFileExtension {
		if "."+ext == pathExtension {
			log.Println("Streaming: [" + pathExtension + "] " + realURLString)
			streamHTTPClientResponceBody(resp, w)
			log.Println("Completed: [" + pathExtension + "] " + realURLString)
			return
		}
	}

	log.Println("Parsing: [" + pathExtension + "] " + realURLString)
	parseHTTPClientResponceBody(resp, w, listName)
	log.Println("Completed: [" + pathExtension + "] " + realURLString)
}

func parseHTTPClientResponceBody(resp *http.Response, w http.ResponseWriter, listName string) {
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		re := regexp.MustCompile(urlRegex)
		urlsToReplace := re.FindAllString(line, -1)
		for _, urlToReplace := range urlsToReplace {
			proxiedURL, err := urlconvert.ConvertURLtoProxyURL(urlToReplace, config.GetConfig().AppURL, listName)
			if err != nil {
				log.Println("Unable to convert url: " + urlToReplace)
			}
			line = strings.ReplaceAll(line, urlToReplace, proxiedURL)
		}

		w.Write([]byte(line + "\n"))
	}
}

func streamHTTPClientResponceBody(resp *http.Response, w http.ResponseWriter) {
	binaryDataChecked := false
	buf := make([]byte, 5*1024) //the chunk size
	reader := bufio.NewReader(resp.Body)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			log.Println(err)
			panic(err)
		}
		if n == 0 {
			break
		}

		//test first chunk if we are really dealing with binary data
		if binaryDataChecked == false {
			if detectNullChar(buf[:n]) == false {
				log.Println("Wont stream non binary data")
				return
			}
			binaryDataChecked = true
		}

		w.Write(buf[:n])
	}
}

func detectNullChar(buf []byte) bool {
	for _, b := range buf {
		if b == 0 {
			return true
		}
	}
	return false
}
