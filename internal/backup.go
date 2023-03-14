package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/deweppro/go-sdk/app"
	"github.com/deweppro/go-sdk/file"
	"github.com/deweppro/go-sdk/log"
	"github.com/deweppro/go-sdk/routine"
	"github.com/deweppro/go-sdk/shell"
	"github.com/deweppro/go-sdk/webutil"
	"github.com/deweppro/goppy/plugins/web"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
)

type Backup struct {
	cli  *webutil.ClientHttp
	conf *Config
	log  log.Logger
}

func NewBackup(cli web.ClientHttp, conf *Config, l log.Logger) *Backup {
	conn := cli.Create(
		webutil.ClientHttpOptionCodec(
			func(in interface{}) (body []byte, contentType string, err error) {
				b, err := json.Marshal(in)
				return b, "", err
			},
			func(code int, contentType string, body []byte, out interface{}) error {
				return json.Unmarshal(body, out)
			},
		),
		webutil.ClientHttpOptionHeaders(
			"Accept", "application/vnd.github+json",
			"X-GitHub-Api-Version", "2022-11-28",
			"Authorization", "Bearer "+conf.ApiKey,
			"type", "all",
		),
	)
	return &Backup{
		cli:  conn,
		conf: conf,
		log:  l,
	}
}

func (v *Backup) Up(ctx app.Context) error {
	routine.Interval(ctx.Context(), time.Hour*6, func(ctx context.Context) {
		v.Dump(ctx)
	})
	return nil
}

func (v *Backup) Down() error {
	return nil
}

func (v *Backup) Dump(ctx context.Context) {
	if !file.Exist(v.conf.Folder) {
		if err := os.MkdirAll(v.conf.Folder, 0755); err != nil {
			v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("create backup folder")
			return
		}
	}

	dbFilename := v.conf.Folder + "/db.yaml"

	if !file.Exist(dbFilename) {
		if err := os.WriteFile(dbFilename, []byte(""), 0644); err != nil {
			v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("create db file")
			return
		}
	}

	db := make(DataModel, 0)
	if err := ReadYaml(dbFilename, &db); err != nil {
		v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("read db")
		return
	}

	for _, owner := range v.conf.Owners {
		result := GitHubResponseModel{}
		err := v.cli.Call(ctx, http.MethodGet,
			fmt.Sprintf("https://api.github.com/search/repositories?q=user:%s", owner), nil, &result)
		if err != nil {
			v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("get repos for [%s]", owner)
			continue
		}

		for _, model := range result.Items {
			db[model.Name] = model.Url
		}
	}

	if err := WriteYaml(dbFilename, &db); err != nil {
		v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("write db")
		return
	}

	sh := shell.New()
	defer sh.Close()
	lw := &LogWriter{Log: v.log}
	sh.SetWriter(lw)
	sh.SetEnv("GIT_SSH_COMMAND", "ssh -i "+v.conf.SshCert)

	for dir, uri := range db {
		path := v.conf.Folder + "/" + dir
		sh.SetDir(path)
		if !file.Exist(path + "/.git") {
			if err := os.MkdirAll(path, 0755); err != nil {
				v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("create repo dir [%s]", dir)
				continue
			}
			if err := sh.CallContext(ctx, fmt.Sprintf("git clone %s .", uri)); err != nil {
				v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("clone new repo [%s]", uri)
			}
			continue
		}

		if err := sh.CallContext(ctx, "git stash && git fetch --all"); err != nil {
			v.log.WithFields(log.Fields{"err": err.Error()}).Errorf("pull repo [%s]", uri)
		}
	}
}

func ReadYaml(filename string, data interface{}) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, data)
}

func WriteYaml(filename string, data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0644)
}

type LogWriter struct {
	Log log.Logger
}

func (v *LogWriter) Write(b []byte) (int, error) {
	v.Log.Infof(string(b))
	return len(b), nil
}
