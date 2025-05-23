package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/andrewheberle/iecbyte"
	"github.com/andrewheberle/simplecommand"
	"github.com/andrewheberle/tus-client/pkg/boltstore"
	"github.com/andrewheberle/tus-client/pkg/jsonstore"
	"github.com/andrewheberle/tus-client/pkg/sqlitestore"
	"github.com/bep/simplecobra"
	tus "github.com/eventials/go-tus"
	"github.com/kirsle/configdir"
	"github.com/schollz/progressbar/v3"
)

type rootCommand struct {
	// flags
	tusUrl        string
	inputFile     string
	storeType     tusStoreType
	storePath     string
	disableResume bool
	noProgress    bool
	quiet         bool
	headers       []string
	chunkSize     iecbyte.Flag

	store tus.Store

	*simplecommand.Command
}

const (
	tusStoreNone = iota
	tusStoreAuto
	tusStoreBolt
	tusStoreJson
	tusStoreSqlite
)

type tusStoreType struct {
	s int
}

func (s tusStoreType) String() string {
	switch s.s {
	case tusStoreNone:
		return "none"
	case tusStoreAuto:
		return "auto"
	case tusStoreBolt:
		return "bolt"
	case tusStoreJson:
		return "json"
	case tusStoreSqlite:
		return "sqlite"
	default:
		return "unsupported"
	}
}

func (s tusStoreType) Type() string {
	return "string"
}

func (s *tusStoreType) Set(value string) error {
	switch value {
	case "", "none":
		s.s = tusStoreNone
	case "auto":
		s.s = tusStoreAuto
	case "bolt":
		s.s = tusStoreBolt
	case "json":
		s.s = tusStoreJson
	case "sqlite":
		s.s = tusStoreSqlite
	default:
		return fmt.Errorf("must be one of \"none\", \"auto\", \"bolt\", \"json\" or \"sqlite\"")
	}

	return nil
}

func (c *rootCommand) Init(cd *simplecobra.Commandeer) error {
	c.Command.Init(cd)
	cmd := cd.CobraCommand
	cmd.Short = "A command line tus client"

	// create app specific config location
	configPath := configdir.LocalConfig(c.Name())
	err := configdir.MakePath(configPath)
	if err != nil {
		return err
	}

	// set default for chunk size based on tus.DefaultConfig()
	c.chunkSize = iecbyte.NewFlag(tus.DefaultConfig().ChunkSize)
	c.storeType = tusStoreType{tusStoreSqlite}

	// command line args
	cmd.Flags().StringVar(&c.tusUrl, "url", "", "tus upload URL")
	cmd.MarkFlagRequired("url")
	cmd.Flags().StringVarP(&c.inputFile, "input", "i", "", "File to upload via tus")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringVar(&c.storePath, "storepath", filepath.Join(configPath, "resume.db"), "Path of store to allow resumable uploads")
	cmd.Flags().Var(&c.storeType, "storetype", "Type of store")
	cmd.Flags().BoolVar(&c.disableResume, "disable-resume", false, "Disable the resumption of uploads (disables the use of the store)")
	cmd.Flags().BoolVar(&c.noProgress, "no-progress", false, "Disable progress bar")
	cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "Disable all output except for errors")
	cmd.Flags().StringArrayVarP(&c.headers, "header", "H", []string{}, "Extra HTTP header(s) to add to request (eg \"Authorization: Bearer TOKEN\")")
	cmd.Flags().Var(&c.chunkSize, "chunksize", "Chunks size for uploads")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	if err := c.Command.PreRun(this, runner); err != nil {
		return err
	}

	// assuming resumable uploads are not disabled, try to set up store
	if !c.disableResume {
		c.setupStore()
	}

	return nil
}

func (c *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
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

	// make sure store is closed (if used)
	if c.store != nil {
		defer c.store.Close()
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

func Execute(args []string) error {
	// set up rootCmd
	rootCmd := &rootCommand{
		Command: simplecommand.New("tus-client", "A command line tus client"),
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

func (c *rootCommand) setupStore() error {
	// set up store
	if c.storePath != "" {
		switch c.storeType.s {
		case tusStoreNone:
			return nil
		case tusStoreAuto:
			if strings.HasSuffix(c.storePath, ".bdb") {
				store, err := boltstore.NewBoltStore(c.storePath)
				if err != nil {
					return err
				}

				c.store = store
			} else if strings.HasSuffix(c.storePath, ".db") {
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
				return fmt.Errorf("storepath must be either a Bolt DB (*.bdb), SQLite database (*.db) or a JSON file (*.json)")
			}
		case tusStoreBolt:
			store, err := boltstore.NewBoltStore(c.storePath)
			if err != nil {
				return err
			}

			c.store = store
		case tusStoreJson:
			store, err := jsonstore.NewJsonStore(c.storePath)
			if err != nil {
				return err
			}

			c.store = store
		case tusStoreSqlite:
			store, err := sqlitestore.NewSqliteStore(c.storePath)
			if err != nil {
				return err
			}

			c.store = store
		default:
			return fmt.Errorf("invalid store type")
		}
	}

	return nil
}
