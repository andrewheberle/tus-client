package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewheberle/tus-client/pkg/sqlitestore"
	"github.com/bep/simplecobra"
	tus "github.com/eventials/go-tus"
	"github.com/schollz/progressbar/v3"
)

type rootCommand struct {
	name string

	// flags
	tusUrl      string
	videoFile   string
	apiToken    string
	db          string
	resume      bool
	chunkSizeMb int

	store tus.Store

	commands []simplecobra.Commander
}

func (c *rootCommand) Name() string {
	return c.name
}

func (c *rootCommand) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	cmd.Short = "A command line TUS client"

	// command line args
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "Tus URL")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&c.videoFile, "input", "i", "", "Video file to upload")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&c.apiToken, "token", "", "API token")
	cmd.MarkFlagRequired("token")
	cmd.Flags().StringVar(&c.db, "db", "", "Database to allow resumable uploads")
	cmd.Flags().BoolVar(&c.resume, "resume", false, "Resume a prior upload")
	cmd.Flags().IntVar(&c.chunkSizeMb, "chunksize", 50, "Chunks size (in MB) for uploads")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	// check that a db path was provided when resume is specified
	if c.resume && c.db == "" {
		return fmt.Errorf("when the resume option is provided the db option must also be provided")
	}

	// set up store
	if c.db != "" {
		store, err := sqlitestore.NewSqliteStore(c.db)
		if err != nil {
			return err
		}

		c.store = store
	}

	// validate chunk size
	if c.chunkSizeMb < 5 {
		return fmt.Errorf("chunksize must be >= 5")
	}

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
		ChunkSize:           int64(c.chunkSizeMb) * 1024 * 1024,
		Resume:              c.resume,
		OverridePatchMethod: false,
		Store:               c.store,
		Header:              headers,
		HttpClient:          nil,
	}

	client, err := tus.NewClient(c.tusUrl, config)
	if err != nil {
		return err
	}

	upload, err := tus.NewUploadFromFile(f)
	if err != nil {
		return err
	}

	uploader, err := client.CreateOrResumeUpload(upload)
	if err != nil {
		return err
	}

	// upload chunk by chunk
	bar := progressbar.Default(100, "uploading")
	for {
		bar.Set64(upload.Progress())
		if upload.Finished() {
			break
		}

		if err := uploader.UploadChunck(); err != nil {
			fmt.Printf("Chunk upload failed: %s\n", err)
			break
		}
	}

	if !upload.Finished() {
		return fmt.Errorf("upload incomplete")
	}

	fmt.Printf("Upload completed succesfully\n")
	return nil
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
