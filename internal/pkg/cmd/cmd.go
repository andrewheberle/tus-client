package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/andrewheberle/tus-client/pkg/iecbyte"
	"github.com/andrewheberle/tus-client/pkg/jsonstore"
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
	storePath     string
	disableResume bool
	// chunkSizeMb   int
	noProgress bool
	quiet      bool
	headers    []string
	chunkSize  iecbyte.Flag

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

	// set default
	c.chunkSize = iecbyte.NewFlag(tus.DefaultConfig().ChunkSize)

	// command line args
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "tus upload URL")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&c.inputFile, "input", "i", "", "File to upload via tus")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&c.storePath, "storepath", filepath.Join(configPath, "resume.db"), "Path of store to allow resumable uploads")
	cmd.Flags().BoolVar(&c.disableResume, "disable-resume", false, "Disable the resumption of uploads (disables the use of the store)")
	// cmd.Flags().IntVar(&c.chunkSizeMb, "chunksize", int(tus.DefaultConfig().ChunkSize/1024/1024), "Chunks size (in MB) for uploads")
	cmd.Flags().BoolVar(&c.noProgress, "no-progress", false, "Disable progress bar")
	cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "Disable all output except for errors")
	cmd.Flags().StringArrayVarP(&c.headers, "header", "H", []string{}, "Extra HTTP header(s) to add to request (eg \"Authorization: Bearer TOKEN\")")
	cmd.Flags().Var(&c.chunkSize, "chunksize", "Chunks size for uploads")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	if !c.disableResume {
		// set up store
		if c.storePath != "" {
			if strings.HasSuffix(c.storePath, ".db") {
				store, err := sqlitestore.NewSqliteStore(c.storePath)
				if err != nil {
					return err
				}

				c.store = store
			} else if strings.HasSuffix(c.storePath, ".json") {
				store, err := jsonstore.NewJsonStore(c.storePath)
				if err != nil {
					return err
				}

				c.store = store
			} else {
				return fmt.Errorf("storepath must be either a SQLite database (*.db) or a JSON file (*.json)")
			}
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

	// create http headers
	headers, err := c.httpHeaders()
	if err != nil {
		return err
	}

	// set up tus config
	config := &tus.Config{
		ChunkSize:           c.chunkSize.Get(),
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

	if !c.quiet {
		fmt.Printf("Starting upload of \"%s\"...", c.inputFile)
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

func (c *rootCommand) httpHeaders() (http.Header, error) {
	// list of tus request headers that cannot be set arbritrarly
	tusHeaders := []string{
		"Upload-Offset",
		"Upload-Length",
		"Tus-Resumable",
		"Upload-Defer-Length",
		"Upload-Metadata",
	}

	headers := make(http.Header)
	for _, h := range c.headers {
		name := http.CanonicalHeaderKey(strings.Split(h, ": ")[0])
		if slices.Contains(tusHeaders, name) {
			return nil, fmt.Errorf("the %s header is used by tus", name)
		}
		value := strings.Join(strings.Split(h, ": ")[1:], ": ")
		headers.Add(name, value)
	}

	return headers, nil
}
