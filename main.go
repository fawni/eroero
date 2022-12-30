package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/adrg/xdg"
	"github.com/cheggaaa/pb/v3"
	"github.com/fawni/eroero/log"
	"github.com/kennygrant/sanitize"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

const baseURL = "https://www.erome.com/"
const baseAlbumURL = baseURL + "a/"

var (
	output     = "."
	configPath = filepath.Join(xdg.ConfigHome, "eroero", "config.json")
	cmd        = &cobra.Command{
		Use:   "eroero <album id>",
		Short: "eroero is a tiny downloader for erome",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			execute(args)
		},
	}
)

func init() {
	termenv.HideCursor()
	cmd.PersistentFlags().StringVarP(&output, "output", "o", output, "output files to a specific directory")
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

	title := doc.Find("h1").First().Text()
	log.Info("Found album \"", title, "\"")
	albumTag := title + " - " + id

	switch output {
	case ".":
		f, err := os.ReadFile(configPath)
		if err != nil {
			break
		}
		cfg := struct {
			Output string `json:"output,omitempty"`
		}{}
		if err := json.Unmarshal(f, &cfg); err != nil {
			break
		}
		output = cfg.Output
	}

	albumOutput := output + "/" + sanitize.Path(albumTag)

	if exists(albumOutput) {
		log.Warn("Album already downloaded.\n")
		redownload := false
		prompt := &survey.Confirm{
			Message: "Download anyway?",
		}
		if err := survey.AskOne(prompt, &redownload); err != nil {
			log.Error(err)
		}
		if !redownload {
			os.Exit(0)
		}
	}

	if err := os.MkdirAll(albumOutput, 0755); err != nil {
		log.Error("Failed to make output directory: ", err)
		os.Exit(1)
	}

	var i, v int
	doc.Find("#album_" + id).First().Find("div > div.media-group").Each(func(_ int, s *goquery.Selection) {
		var url, name string

		switch s.Find("div").HasClass("video") {
		case false:
			url, _ = s.Find(".img").First().Attr("data-src")
			name = filepath.Base(url)
			i++
		default:
			url, _ = s.Find(".video-lg video source").First().Attr("src")
			name = filepath.Base(url)
			v++
		}

		log.Info("Downloading ", name)
		if err := download(url, albumOutput+"/"+name); err != nil {
			log.Error("Failed to download ", name, ": ", err)
		}
	})

	log.Info("Downloaded ", i, " images, ", v, " videos")
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
		log.Warn("File is too big; download might fail.")
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := `{{counters .}} {{bar . " " (green "━") (green "━") (black "━") " "}} {{percent . | magenta}} {{speed . "%s/s" | green}} {{rtime . "%s"| yellow}}`
	bar := pb.
		New64(res.ContentLength).
		SetWidth(80).
		SetTemplateString(tmpl).
		Set(pb.Bytes, true)
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
