package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb"
	"github.com/kennygrant/sanitize"
	"github.com/spf13/cobra"
	"github.com/x6r/eroero/log"
)

const baseURL = "https://www.erome.com/"
const baseAlbumURL = baseURL + "a/"

var (
	output string
	cmd    = &cobra.Command{
		Use:   "eroero <album id>",
		Short: "eroero is a tiny downloader for erome",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			execute(args)
		},
	}
)

func init() {
	cmd.PersistentFlags().StringVarP(&output, "output", "o", ".", "output files to a specific directory")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error(err)
	}
}

func execute(args []string) {
	id := args[0]
	albumURL := baseAlbumURL + id

	log.Info("Fetching media for album ", id, "...")

	res, err := http.Get(albumURL)
	if err != nil {
		log.Error("Failed to find album: ", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Error("Received status error: ", res.Status)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	albumEl := doc.Find("#album_" + id).First()
	var images, videos int

	title := doc.Find("h1").First().Text()
	log.Info("Found album \"", title, "\"")
	albumTag := title + " - " + id

	albumOutput := output + "/" + sanitize.Path(albumTag)

	if exists(albumOutput) {
		log.Warn("Album already downloaded.", "\n")
		redownload := false
		prompt := &survey.Confirm{
			Message: "Download anyway?",
		}
		survey.AskOne(prompt, &redownload)

		if !redownload {
			os.Exit(0)
		}
	}

	if err := os.MkdirAll(albumOutput, 0755); err != nil {
		log.Error("Failed to make output directory: ", err)
		os.Exit(1)
	}

	albumEl.Find("div > div.media-group").Each(func(_ int, s *goquery.Selection) {
		var url, name string
		if s.Find("div").HasClass("video") {
			url, _ = s.Find(".video-lg video source").First().Attr("src")
			name = filepath.Base(url)
			videos++
		} else {
			url, _ = s.Find(".img").First().Attr("data-src")
			name = filepath.Base(url)
			images++
		}

		if url != "" {
			log.Info("Downloading ", name)
			if err := download(url, albumOutput+"/"+name); err != nil {
				log.Error("Failed to download ", name, ": ", err)
			}
		}
	})

	log.Info("Downloaded ", images, " images, ", videos, " videos")
}

func download(url, name string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, strings.NewReader(""))

	if err != nil {
		return err
	}
	req.Header.Add("Referer", baseURL)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.ContentLength > 1e7 {
		log.Warn("File is too big. Download might fail.")
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	bar := pb.New64(res.ContentLength).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowTimeLeft = true
	barWriter := bar.NewProxyReader(res.Body)

	bar.Start()
	_, err = io.Copy(file, barWriter)
	if err != nil {
		return err
	}
	bar.Finish()

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
