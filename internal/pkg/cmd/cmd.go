package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andrewheberle/tus-client/pkg/sqlitestore"
	"github.com/bep/simplecobra"
	tus "github.com/eventials/go-tus"
	"github.com/kirsle/configdir"
	"github.com/schollz/progressbar/v3"
)

type rootCommand struct {
	name string

	// flags
	tusUrl        string
	videoFile     string
	bearerToken   string
	db            string
	disableResume bool
	chunkSizeMb   int

	store tus.Store

	commands []simplecobra.Commander
}

func (c *rootCommand) Name() string {
	return c.name
}

func (c *rootCommand) Init(cd *simplecobra.Commandeer) error {
	cmd := cd.CobraCommand
	cmd.Short = "A command line tus client"

	// create app specific config location
	configPath := configdir.LocalConfig(c.Name())
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}

	// command line args
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "tus URL")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&c.videoFile, "input", "i", "", "Video file to upload")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&c.bearerToken, "token", "", "Authorization Bearer token")
	cmd.Flags().StringVar(&c.db, "db", filepath.Join(configPath, "resume.db"), "Database to allow resumable uploads")
	cmd.Flags().BoolVar(&c.disableResume, "disable-resume", false, "Disable the resumption of uploads (disables the use of the database)")
	cmd.Flags().IntVar(&c.chunkSizeMb, "chunksize", 5, "Chunks size (in MB) for uploads")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	if !c.disableResume {
		// set up store
		if c.db != "" {
			store, err := sqlitestore.NewSqliteStore(c.db)
			if err != nil {
				return err
			}

			c.store = store
		}
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
	if c.bearerToken != "" {
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	}

	config := &tus.Config{
		ChunkSize:           int64(c.chunkSizeMb) * 1024 * 1024,
		Resume:              !c.disableResume,
		OverridePatchMethod: false,
		Store:               c.store,
		Header:              headers,
		HttpClient:          nil,
	}

	// initialise new tus client
	client, err := tus.NewClient(c.tusUrl, config)
	if err != nil {
		return err
	}

	// create upload from file
	upload, err := tus.NewUploadFromFile(f)
	if err != nil {
		return err
	}

	// create uploader or resume if partial from upload
	uploader, err := client.CreateOrResumeUpload(upload)
	if err != nil {
		return err
	}

	// set up progress bar
	bar := progressbar.DefaultBytes(upload.Size(), "uploading")
	defer bar.Close()

	// set initial upload progress
	bar.Set64(upload.Offset())
	for {
		// check if upload is done
		if upload.Finished() {
			break
		}

		// upload next chunk
		if err := uploader.UploadChunck(); err != nil {
			fmt.Printf("Chunk upload failed: %s\n", err)
			break
		}

		// update progress
		bar.Set64(upload.Offset())
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
