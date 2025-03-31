package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	inputFile     string
	db            string
	disableResume bool
	chunkSizeMb   int
	noProgress    bool
	quiet         bool
	headers       []string

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
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "tus upload URL")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&c.inputFile, "input", "i", "", "File to upload via tus")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&c.db, "db", filepath.Join(configPath, "resume.db"), "Path of database to allow resumable uploads")
	cmd.Flags().BoolVar(&c.disableResume, "disable-resume", false, "Disable the resumption of uploads (disables the use of the database)")
	cmd.Flags().IntVar(&c.chunkSizeMb, "chunksize", 10, "Chunks size (in MB) for uploads")
	cmd.Flags().BoolVar(&c.noProgress, "no-progress", false, "Disable progress bar")
	cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "Disable all output except for errors")
	cmd.Flags().StringArrayVarP(&c.headers, "header", "H", []string{}, "Extra HTTP header(s) to add to request (eg \"Authorization: Bearer TOKEN\")")

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

	return nil
}

func (c *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	// make sure store is closed
	if c.store != nil {
		defer c.store.Close()
	}

	// open input file
	f, err := os.Open(c.inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	headers := make(http.Header)
	for _, h := range c.headers {
		n := strings.Split(h, ": ")[0]
		v := strings.Join(strings.Split(h, ": ")[1:], ": ")
		headers.Add(n, v)
	}

	// set up tus config
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
	var bar *progressbar.ProgressBar
	if c.noProgress || c.quiet {
		bar = progressbar.DefaultBytesSilent(upload.Size(), "uploading")
	} else {
		bar = progressbar.DefaultBytes(upload.Size(), "uploading")
	}
	defer bar.Close()

	// set initial upload progress
	bar.Set64(upload.Offset())

	if !c.quiet {
		fmt.Printf("Starting upload of \"%s\"...", c.inputFile)
	}
	for {
		// check if upload is done
		if upload.Finished() {
			break
		}

		// upload next chunk
		if err := uploader.UploadChunck(); err != nil {
			return fmt.Errorf("chunk upload failed: %w", err)
		}

		// update progress
		bar.Set64(upload.Offset())
	}

	if !upload.Finished() {
		return fmt.Errorf("upload incomplete")
	}

	if !c.quiet {
		fmt.Printf("Upload completed successfully\n")
	}
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
