package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bep/simplecobra"
	tus "github.com/eventials/go-tus"
)

type rootCommand struct {
	name string

	// flags
	tusUrl    string
	videoFile string
	apiToken  string

	commands []simplecobra.Commander
}

func (c *rootCommand) Name() string {
	return c.name
}

func (c *rootCommand) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	cmd.Short = "A simple HTTP server"

	// command line args
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "Tus URL")
	cmd.Flags().StringVarP(&c.videoFile, "input", "i", "", "Video file to upload")
	cmd.Flags().StringVar(&c.apiToken, "token", "", "API token")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	return nil
}

func (c *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	f, err := os.Open(c.videoFile)

	if err != nil {
		return err
	}

	defer f.Close()

	headers := make(http.Header)
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	config := &tus.Config{
		ChunkSize:           50 * 1024 * 1024, // Required a minimum chunk size of 5 MB, here we use 50 MB.
		Resume:              false,
		OverridePatchMethod: false,
		Store:               nil,
		Header:              headers,
		HttpClient:          nil,
	}

	client, _ := tus.NewClient(c.tusUrl, config)

	upload, _ := tus.NewUploadFromFile(f)

	uploader, _ := client.CreateUpload(upload)

	return uploader.Upload()
}

func (c *rootCommand) Commands() []simplecobra.Commander {
	return c.commands
}

func Execute(args []string) error {
	// set up rootCmd
	rootCmd := &rootCommand{
		name: "tus-client",
	}
	x, err := simplecobra.New(rootCmd)
	if err != nil {
		return err
	}

	if _, err := x.Execute(context.Background(), args); err != nil {
		return err
	}

	return nil
}
